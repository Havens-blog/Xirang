#Requires -Version 5.1
<#
.SYNOPSIS
  Haven ops platform - local one-shot deploy (Windows, no Docker: bare processes + nginx)

.DESCRIPTION
  Orchestration-only repo. The actual source code lives in the sibling repos under
  D:\Haven (eiam / e-cam-service / opsagent / haven-console / opsagent-web / e-cam-web).
  This script builds and runs them; it does NOT own any application source code.

  1. Build the three Go backends to D:\Haven\logs\dev\bin
     (eiam / e-cam-service / opsagent)
  2. Build the two static frontends (haven-console, opsagent-web; served by nginx)
  3. Start missing services by free-port check (never kills existing processes),
     then reload nginx
  4. e-cam-web runs as vite dev server: not started unless -WithCamWeb

  Port contract (aligned with D:\Haven\nginx-1.25.4\conf\vhost\nginx.dev.conf):
    nginx :8888 -> eiam :9000 / e-cam-service :8001 / opsagent :8081 /
                    /cam/ -> :5173 dev / /console/ /opsagent/ -> static dist
  Ports already in use are left alone.

  NOTE: keep this file ASCII-only. Windows PowerShell 5.1 reads BOM-less .ps1 as
  the ANSI codepage (GBK); non-ASCII comments break tokenization after any edit.
#>
[CmdletBinding()]
param(
  [string]$Workspace     = 'D:\Haven',                       # where all sibling repos live
  [string]$BinDir        = 'D:\Haven\logs\dev\bin',          # Go build output
  [string]$NginxExe      = 'D:\Haven\nginx-1.25.4\nginx.exe',
  [string]$NginxPrefix   = 'D:\Haven\nginx-1.25.4\',         # -s reload resolves conf vs CWD on Windows
  [string]$OpsAgentConfig = 'D:\Haven\opsagent\config',      # port-8081 archive config, matches nginx
  [string]$EcamConfig    = 'config\prod.yaml',               # relative to e-cam-service repo root
  [switch]$SkipBackendBuild,
  [switch]$SkipFrontendBuild,
  [switch]$SkipStart,
  [switch]$WithCamWeb,
  [switch]$Restart
)

$ErrorActionPreference = 'Stop'

# Source-of-truth repo roots (code lives here, NOT in this orchestration repo)
$eiamRepo    = Join-Path $Workspace 'eiam'
$ecamRepo    = Join-Path $Workspace 'e-cam-service'
$opsRepo     = Join-Path $Workspace 'opsagent'
$consoleRepo = Join-Path $Workspace 'haven-console'
$opsWebRepo  = Join-Path $Workspace 'opsagent-web'
$camWebRepo  = Join-Path $Workspace 'e-cam-web'

if ($Restart) {
  info 'restart mode: stopping dev-managed services first (other instances untouched)'
  & (Join-Path $PSScriptRoot 'dev-down.ps1')
}

function info($m)  { Write-Host "==> $m" -ForegroundColor Cyan }
function ok($m)    { Write-Host " OK  $m" -ForegroundColor Green }
function warn($m)  { Write-Host " !!  $m" -ForegroundColor Yellow }
function die($m)   { Write-Host " XX  $m" -ForegroundColor Red; exit 1 }

function Test-PortFree([int]$Port) {
  $null -eq (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue)
}

function Start-AppProc([int]$Port, [string]$Name, [string]$Exe, [string]$ProcArgs, [string]$Cwd) {
  if (-not (Test-PortFree $Port)) {
    warn "$Name :$Port is busy - skipping (existing instance reused)"
    return
  }
  # LONGEVITY: with -RedirectStandardOutput/-RedirectStandardError, Start-Process
  # disables UseShellExecute and the child INHERITS this shell's console; when the
  # launcher console dies the child gets CTRL_CLOSE_EVENT and exits silently.
  # Detached recipe: hide the window, do NOT redirect -> child keeps its own console.
  Start-Process -FilePath $Exe -ArgumentList $ProcArgs -WorkingDirectory $Cwd -WindowStyle Hidden
  $deadline = (Get-Date).AddSeconds(30)
  while ((Test-PortFree $Port) -and (Get-Date) -lt $deadline) { Start-Sleep -Milliseconds 500 }
  if (Test-PortFree $Port) { warn "$Name did not listen on :$Port within 30s - see logs under $Cwd" }
  else { ok "$Name listening on :$Port" }
}

$env:CGO_ENABLED = '0'

# ------------------------- 1. backend builds (from old repos) -------------------------
if ($SkipBackendBuild) { info 'backend build skipped' }
else {
  info 'building 3 Go backends -> D:\Haven\logs\dev\bin'
  New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

  Push-Location $eiamRepo
  go build -o (Join-Path $BinDir 'eiam.exe') .
  if ($LASTEXITCODE -ne 0) { Pop-Location; die 'eiam build failed' }
  Pop-Location

  Push-Location $ecamRepo
  go build -o (Join-Path $BinDir 'e-cam-service.exe') .
  if ($LASTEXITCODE -ne 0) { Pop-Location; die 'e-cam-service build failed' }
  Pop-Location

  # opsagent is NOT a member of D:\Haven\go.work (it resolves deps standalone),
  # so disable the workspace to avoid "module is not one of the workspace modules".
  Push-Location $opsRepo
  $prevGowork = $env:GOWORK
  $env:GOWORK = 'off'
  go build -o (Join-Path $BinDir 'opsagent.exe') ./cmd/opsagent
  $buildExit = $LASTEXITCODE
  if ($prevGowork) { $env:GOWORK = $prevGowork } else { Remove-Item Env:GOWORK -ErrorAction SilentlyContinue }
  if ($buildExit -ne 0) { Pop-Location; die 'opsagent build failed' }
  Pop-Location
  ok 'backend builds done'
}

# ------------------------- 2. frontend builds (from old repos) -------------------------
if ($SkipFrontendBuild) { info 'frontend build skipped' }
else {
  foreach ($dir in @($consoleRepo, $opsWebRepo)) {
    $app = Split-Path -Leaf $dir
    info "building frontend $app"
    Push-Location $dir
    if (-not (Test-Path 'node_modules')) { pnpm install --no-frozen-lockfile }
    pnpm build
    if ($LASTEXITCODE -ne 0) { Pop-Location; die "$app build failed" }
    Pop-Location
  }
  ok 'frontend builds done (dist served by nginx)'
}

if ($SkipStart) { info 'start skipped (build only)'; return }

# ------------------------- 3. start backends -------------------------
info 'starting backends (free ports only)'
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

# eiam: config/config.yaml is resolved against the WORKING DIRECTORY - cwd = repo root
Start-AppProc 9000 'eiam' (Join-Path $BinDir 'eiam.exe') 'server' $eiamRepo

# e-cam-service: EGO_CONFIG_PATH relative to module root (prod.yaml = remote DB + cloud creds)
$env:EGO_CONFIG_PATH = $EcamConfig
Start-AppProc 8001 'e-cam-service' (Join-Path $BinDir 'e-cam-service.exe') 'start' $ecamRepo
Remove-Item Env:EGO_CONFIG_PATH -ErrorAction SilentlyContinue

# opsagent: -config points at a credential-bearing config dir; expects a LOCAL mongod
# on 127.0.0.1:27017 and the shared remote redis.
if (-not (Test-Path $OpsAgentConfig)) {
  warn "opsagent config dir missing: $OpsAgentConfig - skipped"
} else {
  Start-AppProc 8081 'opsagent' (Join-Path $BinDir 'opsagent.exe') "-config `"$OpsAgentConfig`"" $opsRepo
}

# ------------------------- 4. cam-web dev server (optional) -------------------------
if ($WithCamWeb) {
  if (-not (Test-PortFree 5173)) { warn 'e-cam-web :5173 already running (reusing existing dev server)' }
  else {
    info 'starting e-cam-web vite dev server (:5173)'
    Start-Process -FilePath 'pnpm.cmd' -ArgumentList 'dev' -WorkingDirectory $camWebRepo -WindowStyle Hidden
  }
} else {
  warn 'e-cam-web not managed by this script; for nginx /cam/ run: cd D:\Haven\e-cam-web; pnpm dev (or pass -WithCamWeb)'
}

# ------------------------- 5. nginx -------------------------
info 'nginx (:8888 unified entry)'
if (Test-Path $NginxExe) {
  if (Get-Process nginx -ErrorAction SilentlyContinue) { & $NginxExe -p $NginxPrefix -s reload; ok 'reloaded running nginx' }
  else { Start-Process -FilePath $NginxExe -ArgumentList "-p `"$NginxPrefix`"" -WorkingDirectory $NginxPrefix; ok 'started nginx :8888' }
} else { warn "nginx not found: $NginxExe (skipped; reload manually)" }

# ------------------------- 6. summary -------------------------
Write-Host ''
Write-Host '================ local deploy summary ================' -ForegroundColor Cyan
Write-Host '  console        http://localhost:8888/console/'
Write-Host '  cam frontend   http://localhost:8888/cam/       (needs e-cam-web dev :5173)'
Write-Host '  agent frontend http://localhost:8888/opsagent/'
Write-Host '  API            /api/iam/* -> :9000 ; /api/v1/{cam,cmdb,certs} -> :8001 ; /api/v1/opsagent/ -> :8081'
Write-Host '=====================================================' -ForegroundColor Cyan
