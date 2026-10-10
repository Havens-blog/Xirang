#Requires -Version 5.1
<#
.SYNOPSIS
  Haven ops platform - local one-shot deploy (Windows, no Docker: bare processes + nginx)

.DESCRIPTION
  1. Build the three Go backends to logs\dev\bin (eiam / e-cam-service / opsagent-api)
  2. Build the two static frontends (haven-console, opsagent-web; served by nginx)
  3. Start missing services by free-port check (never kills existing processes),
     then reload nginx
  4. e-cam-web runs as vite dev server: not started unless -WithCamWeb

  Port contract (aligned with nginx-1.25.4\conf\vhost\nginx.dev.conf):
    nginx :8888 -> eiam :9000 / e-cam-service :8001 / opsagent :8081 /
                    /cam/ -> :5173 dev / /console/ /opsagent/ -> static dist
  Ports already in use are left alone - old and new lines can coexist.

  NOTE: keep this file ASCII-only. Windows PowerShell 5.1 reads BOM-less .ps1 as
  the ANSI codepage (GBK); non-ASCII comments break tokenization after any edit.
#>
[CmdletBinding()]
param(
  [string]$NginxExe       = 'D:\Haven\nginx-1.25.4\nginx.exe',
  [string]$NginxPrefix    = 'D:\Haven\nginx-1.25.4\',   # -s reload resolves conf vs CWD on Windows; pin the prefix
  [string]$OpsAgentConfig = 'D:\Haven\opsagent\config',   # port-8081 archive config, matches nginx
  [string]$EcamConfig     = 'config\prod.yaml',            # relative to services\e-cam-service
  [switch]$SkipBackendBuild,
  [switch]$SkipFrontendBuild,
  [switch]$SkipStart,
  [switch]$WithCamWeb,
  [switch]$Restart
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot          # repo root
$bin  = Join-Path $root 'logs\dev\bin'

if ($Restart) {
  info 'restart mode: stopping mono-managed services first (old-line instances untouched)'
  & (Join-Path $PSScriptRoot 'dev-down.ps1')
}

function info($m)  { Write-Host "==> $m" -ForegroundColor Cyan }
function ok($m)    { Write-Host " OK  $m" -ForegroundColor Green }
function warn($m)  { Write-Host " !!  $m" -ForegroundColor Yellow }
function die($m)   { Write-Host " XX  $m" -ForegroundColor Red; exit 1 }

function Test-PortFree([int]$Port) {
  $null -eq (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue)
}

function Start-AppProc([int]$Port, [string]$Name, [string]$Exe, [string]$ProcArgs, [string]$Cwd, [string]$LogBase) {
  if (-not (Test-PortFree $Port)) {
    warn "$Name :$Port is busy - skipping (existing instance reused)"
    return
  }
  Start-Process -FilePath $Exe -ArgumentList $ProcArgs -WorkingDirectory $Cwd `
    -RedirectStandardOutput "$LogBase.out.log" -RedirectStandardError "$LogBase.err.log" | Out-Null
  $deadline = (Get-Date).AddSeconds(30)
  while ((Test-PortFree $Port) -and (Get-Date) -lt $deadline) { Start-Sleep -Milliseconds 500 }
  if (Test-PortFree $Port) { warn "$Name did not listen on :$Port within 30s - see $LogBase.err.log" }
  else { ok "$Name listening on :$Port" }
}

$env:CGO_ENABLED = '0'

# ------------------------- 1. backend builds -------------------------
if ($SkipBackendBuild) { info 'backend build skipped' }
else {
  info 'building 3 Go backends -> logs\dev\bin'
  New-Item -ItemType Directory -Force -Path $bin | Out-Null

  Push-Location (Join-Path $root 'services\eiam')
  go build -o (Join-Path $bin 'eiam.exe') .
  if ($LASTEXITCODE -ne 0) { Pop-Location; die 'eiam build failed' }
  Pop-Location

  Push-Location (Join-Path $root 'services\e-cam-service')
  go build -o (Join-Path $bin 'e-cam-service.exe') .
  if ($LASTEXITCODE -ne 0) { Pop-Location; die 'e-cam-service build failed' }
  Pop-Location

  Push-Location (Join-Path $root 'services\opsagent')
  go build -o (Join-Path $bin 'opsagent-api.exe') ./cmd/opsagent
  if ($LASTEXITCODE -ne 0) { Pop-Location; die 'opsagent build failed' }
  Pop-Location
  ok 'backend builds done'
}

# ------------------------- 2. frontend builds -------------------------
if ($SkipFrontendBuild) { info 'frontend build skipped' }
else {
  foreach ($app in @('haven-console','opsagent-web')) {
    $dir = Join-Path $root "apps\$app"
    info "building frontend $app"
    Push-Location $dir
    if (-not (Test-Path 'node_modules')) { pnpm install --no-frozen-lockfile }
    # prefer build:gz (gzip budget chain); fall back to plain build on old snapshots
    $pkg = ([System.IO.File]::ReadAllText((Join-Path $dir 'package.json'))) | ConvertFrom-Json
    if ($pkg.scripts.'build:gz') { pnpm build:gz } else { pnpm build }
    Pop-Location
  }
  ok 'frontend builds done (dist served by nginx)'
}

if ($SkipStart) { info 'start skipped (build only)'; return }

# ------------------------- 3. start backends -------------------------
info 'starting backends (free ports only)'
New-Item -ItemType Directory -Force -Path $bin | Out-Null

# eiam: config/config.yaml is resolved against the WORKING DIRECTORY - cwd must be the module root
Start-AppProc 9000 'eiam' (Join-Path $bin 'eiam.exe') 'server' `
  (Join-Path $root 'services\eiam') (Join-Path $bin 'eiam')

# e-cam-service: EGO_CONFIG_PATH relative to module root (prod.yaml = remote DB + cloud creds)
$env:EGO_CONFIG_PATH = $EcamConfig
Start-AppProc 8001 'e-cam-service' (Join-Path $bin 'e-cam-service.exe') 'start' `
  (Join-Path $root 'services\e-cam-service') (Join-Path $bin 'e-cam-service')
Remove-Item Env:EGO_CONFIG_PATH -ErrorAction SilentlyContinue

# opsagent: -config points at a credential-bearing config dir; archive config uses port 8081 = nginx.
# NOTE: this config expects a LOCAL mongod on 127.0.0.1:27017 (not installed on this box) and the
# shared remote redis. If mongod is absent the app exits - see logs\dev\bin\opsagent-api.err.log.
if (-not (Test-Path $OpsAgentConfig)) {
  warn "opsagent config dir missing: $OpsAgentConfig - skipped (prepare a config dir with llm creds etc. first)"
} else {
  Start-AppProc 8081 'opsagent-api' (Join-Path $bin 'opsagent-api.exe') "-config `"$OpsAgentConfig`"" `
    (Join-Path $root 'services\opsagent') (Join-Path $bin 'opsagent-api')
}

# ------------------------- 4. cam-web dev server (optional) -------------------------
if ($WithCamWeb) {
  if (-not (Test-PortFree 5173)) { warn 'e-cam-web :5173 already running (reusing existing dev server)' }
  else {
    info 'starting e-cam-web vite dev server (:5173)'
    Start-Process -FilePath 'cmd.exe' -ArgumentList '/c pnpm dev' `
      -WorkingDirectory (Join-Path $root 'apps\e-cam-web') `
      -RedirectStandardOutput (Join-Path $bin 'e-cam-web.out.log') `
      -RedirectStandardError (Join-Path $bin 'e-cam-web.err.log') | Out-Null
  }
} else {
  warn 'e-cam-web not managed by this script; for nginx /cam/ run: cd apps\e-cam-web; pnpm dev (or pass -WithCamWeb)'
}

# ------------------------- 5. nginx -------------------------
info 'nginx (:8888 unified entry)'
if (Test-Path $NginxExe) {
  if (Get-Process nginx -ErrorAction SilentlyContinue) { & $NginxExe -p $NginxPrefix -s reload; ok 'reloaded running nginx' }
  else { Start-Process -FilePath $NginxExe -ArgumentList "-p `"$NginxPrefix`"" -WorkingDirectory $NginxPrefix; ok 'started nginx :8888' }
} else { warn "nginx not found: $NginxExe (skipped; reload manually with: nginx -p<$NginxPrefix> -s reload)" }

# ------------------------- 6. summary -------------------------
Write-Host ''
Write-Host '================ local deploy summary ================' -ForegroundColor Cyan
Write-Host '  console       http://localhost:8888/console/'
Write-Host '  cam frontend  http://localhost:8888/cam/       (needs e-cam-web dev :5173)'
Write-Host '  agent frontend http://localhost:8888/opsagent/'
Write-Host '  API           /api/iam/* -> :9000 ; /api/v1/{cam,cmdb,certs} -> :8001 ; /api/v1/opsagent/ -> :8081'
Write-Host '=====================================================' -ForegroundColor Cyan