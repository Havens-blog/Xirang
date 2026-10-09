#Requires -Version 5.1
<#
.SYNOPSIS
  息壤运维平台 · 本地一键部署（Windows 无 Docker · 裸进程 + nginx 形态）

.DESCRIPTION
  1. 构建三个 Go 后端 → logs\dev\bin\（eiam / e-cam-service / opsagent-api）
  2. 构建两个静态前端 dist（haven-console、opsagent-web；由 nginx 直接托管）
  3. 按端口占用自动跳过已在跑的实例，启动缺的进程，最后 reload nginx
  4. e-cam-web 走 vite dev server（默认只提示；-WithCamWeb 由本脚本托管）

  端口契约（与 nginx-1.25.4\conf\vhost\nginx.dev.conf 对齐）：
    nginx :8888 → eiam :9000 / e-cam-service :8001 / opsagent :8081 /
                 /cam/ → :5173 dev / /console/ /opsagent/ → dist
  已在跑的端口不杀进程，新旧两线可并存（适合灰度切换）。
#>
[CmdletBinding()]
param(
  [string]$NginxExe       = 'D:\Haven\nginx-1.25.4\nginx.exe',
  [string]$OpsAgentConfig = 'D:\Haven\opsagent\config',   # 端口 8081 的归档配置（与 nginx 对齐）
  [string]$EcamConfig     = 'config\prod.yaml',            # 相对 services\e-cam-service
  [switch]$SkipBackendBuild,
  [switch]$SkipFrontendBuild,
  [switch]$SkipStart,
  [switch]$WithCamWeb
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot          # 仓库根
$bin  = Join-Path $root 'logs\dev\bin'

function info($m)  { Write-Host "==> $m" -ForegroundColor Cyan }
function ok($m)    { Write-Host " ✓ $m" -ForegroundColor Green }
function warn($m)  { Write-Host " ⚠ $m" -ForegroundColor Yellow }
function die($m)   { Write-Host " ✗ $m" -ForegroundColor Red; exit 1 }

function Test-PortFree([int]$Port) {
  $null -eq (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue)
}

function Start-Service([int]$Port, [string]$Name, [string]$Exe, [string]$Args, [string]$Cwd, [string]$LogBase) {
  if (-not (Test-PortFree $Port)) {
    warn "$Name :$Port 已被占用 → 跳过启动（沿用现役实例）"
    return
  }
  Start-Process -FilePath $Exe -ArgumentList $Args -WorkingDirectory $Cwd `
    -RedirectStandardOutput "$LogBase.out.log" -RedirectStandardError "$LogBase.err.log" | Out-Null
  $deadline = (Get-Date).AddSeconds(30)
  while ((Test-PortFree $Port) -and (Get-Date) -lt $deadline) { Start-Sleep -Milliseconds 500 }
  if (Test-PortFree $Port) { warn "$Name 启动后 :$Port 未监听（30s），看 $LogBase.err.log" }
  else { ok "$Name 已监听 :$Port" }
}

$env:CGO_ENABLED = '0'

# ───────────────────────── 1. 后端构建 ─────────────────────────
if ($SkipBackendBuild) { info '跳过后端构建' }
else {
  info '构建三个 Go 后端 → logs\dev\bin'
  New-Item -ItemType Directory -Force -Path $bin | Out-Null

  Push-Location (Join-Path $root 'services\eiam')
  go build -o (Join-Path $bin 'eiam.exe') .
  Pop-Location

  Push-Location (Join-Path $root 'services\e-cam-service')
  go build -o (Join-Path $bin 'e-cam-service.exe') .
  Pop-Location

  Push-Location (Join-Path $root 'services\opsagent')
  go build -o (Join-Path $bin 'opsagent-api.exe') ./cmd/opsagent
  Pop-Location
  ok '后端构建完成'
}

# ───────────────────────── 2. 前端构建 ─────────────────────────
if ($SkipFrontendBuild) { info '跳过前端构建' }
else {
  foreach ($app in @('haven-console','opsagent-web')) {
    $dir = Join-Path $root "apps\$app"
    info "构建前端 $app"
    Push-Location $dir
    if (-not (Test-Path 'node_modules')) { pnpm install --no-frozen-lockfile }
    pnpm build:gz
    Pop-Location
  }
  ok '前端构建完成（dist 由 nginx 托管）'
}

if ($SkipStart) { info '跳过启动（仅构建）'; return }

# ───────────────────────── 3. 启动后端 ─────────────────────────
info '启动后端（端口空闲才启动）'
New-Item -ItemType Directory -Force -Path $bin | Out-Null

# eiam：配置按【工作目录】解析 config/config.yaml（模块根内，勿改）
Start-Service 9000 'eiam' (Join-Path $bin 'eiam.exe') 'server' `
  (Join-Path $root 'services\eiam') (Join-Path $bin 'eiam')

# e-cam-service：EGO_CONFIG_PATH 相对模块根（prod.yaml = 远程库+云凭据口径）
$env:EGO_CONFIG_PATH = $EcamConfig
Start-Service 8001 'e-cam-service' (Join-Path $bin 'e-cam-service.exe') 'start' `
  (Join-Path $root 'services\e-cam-service') (Join-Path $bin 'e-cam-service')
Remove-Item Env:EGO_CONFIG_PATH -ErrorAction SilentlyContinue

# opsagent：-config 指向含密钥的配置目录；归档配置端口 8081 与 nginx 对齐
if (-not (Test-Path $OpsAgentConfig)) {
  warn "opsagent 配置目录不存在：$OpsAgentConfig → 跳过（准备含 llm 等密钥的 config 目录后再跑）"
} else {
  Start-Service 8081 'opsagent-api' (Join-Path $bin 'opsagent-api.exe') "-config `"$OpsAgentConfig`"" `
    (Join-Path $root 'services\opsagent') (Join-Path $bin 'opsagent-api')
}

# ───────────────────────── 4. 前端 dev server（可选） ─────────────────────────
if ($WithCamWeb) {
  if (-not (Test-PortFree 5173)) { warn 'e-cam-web :5173 已在跑（现役 dev server 直接复用）' }
  else {
    info '启动 e-cam-web vite dev server（:5173）'
    Start-Process -FilePath 'cmd.exe' -ArgumentList '/c pnpm dev' `
      -WorkingDirectory (Join-Path $root 'apps\e-cam-web') `
      -RedirectStandardOutput (Join-Path $bin 'e-cam-web.out.log') `
      -RedirectStandardError (Join-Path $bin 'e-cam-web.err.log') | Out-Null
  }
} else {
  warn 'e-cam-web 未由脚本托管；如需经 nginx /cam/ 访问，请自行：cd apps\e-cam-web && pnpm dev（或加 -WithCamWeb）'
}

# ───────────────────────── 5. nginx ─────────────────────────
info 'nginx（:8888 统一入口）'
if (Test-Path $NginxExe) {
  if (Get-Process nginx -ErrorAction SilentlyContinue) { & $NginxExe -s reload; ok '已 reload 现役 nginx' }
  else { Start-Process -FilePath $NginxExe; ok '已启动 nginx :8888' }
} else { warn "nginx 未找到：$NginxExe（跳过；手改 conf 后自行 nginx -s reload）" }

# ───────────────────────── 6. 收尾 ─────────────────────────
Write-Host ''
Write-Host '════════ 本地部署入口 ════════' -ForegroundColor Cyan
Write-Host '  控制台    http://localhost:8888/console/'
Write-Host '  云管前台  http://localhost:8888/cam/       （需要 e-cam-web dev :5173）'
Write-Host '  Agent前端 http://localhost:8888/opsagent/'
Write-Host '  API       /api/iam/* → :9000 · /api/v1/{cam,cmdb,certs} → :8001 · /api/v1/opsagent/ → :8081'
Write-Host '════════════════════════════' -ForegroundColor Cyan