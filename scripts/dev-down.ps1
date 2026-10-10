#Requires -Version 5.1
<#
.SYNOPSIS
  Stop the services that dev-up.ps1 started, and ONLY those.

.DESCRIPTION
  Orchestration-only repo: services are built from the old repos into
  D:\Haven\logs\dev\bin. For ports 9000 / 8001 / 8081, look at the listening
  process and kill it only when its executable lives under that bin dir.
  Other instances, the vite dev server and nginx are left alone unless
  explicitly requested via -IncludeCamWeb / -IncludeNginx.

  Restart pattern:  .\scripts\dev-up.ps1 -Restart

  NOTE: keep this file ASCII-only (see dev-up.ps1 header note).
#>
[CmdletBinding(SupportsShouldProcess)]
param(
  [string]$BinDir = 'D:\Haven\logs\dev\bin',
  [string]$CamWebRepo = 'D:\Haven\e-cam-web',
  [switch]$IncludeCamWeb,   # also stop the e-cam-web vite dev (:5173) if it is ours
  [switch]$IncludeNginx     # also stop nginx gracefully (-s stop); OFF by default
)

$ErrorActionPreference = 'Stop'

$map = @{ 9000 = 'eiam.exe'; 8001 = 'e-cam-service.exe'; 8081 = 'opsagent.exe' }

function Get-ListenPids([int]$Port) {
  (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique)
}

foreach ($port in ($map.Keys | Sort-Object)) {
  $want = (Join-Path $BinDir $map[$port]).ToLowerInvariant()
  $pids = @(Get-ListenPids $port)
  if ($pids.Count -eq 0) { Write-Host "port ${port}: free (nothing to stop)" -ForegroundColor DarkGray; continue }
  foreach ($pid_ in $pids) {
    $proc = Get-Process -Id $pid_ -ErrorAction SilentlyContinue
    if (-not $proc) { Write-Host "port ${port}: PID $pid_ already gone" -ForegroundColor DarkGray; continue }
    $path = $proc.Path
    if ($path -and $path.ToLowerInvariant() -eq $want) {
      if ($PSCmdlet.ShouldProcess("$($proc.ProcessName) PID $pid_ on :$port", 'Stop dev-managed service')) {
        Stop-Process -Id $pid_ -Force
        Write-Host " [stopped] $($proc.ProcessName) PID $pid_ on :$port (dev build)" -ForegroundColor Green
      }
    } else {
      Write-Host " [kept]    port ${port} owned by $($proc.ProcessName) PID $pid_ (path: $path) - not a dev build, left alone" -ForegroundColor Yellow
    }
  }
}

if ($IncludeCamWeb) {
  $pids = @(Get-ListenPids 5173)
  foreach ($pid_ in $pids) {
    $cli = (Get-CimInstance Win32_Process -Filter "ProcessId=$pid_" -ErrorAction SilentlyContinue).CommandLine
    if ($cli -and $cli.ToLowerInvariant().Contains($CamWebRepo.ToLowerInvariant())) {
      if ($PSCmdlet.ShouldProcess("vite dev PID $pid_ on :5173 (e-cam-web)", 'Stop')) {
        Stop-Process -Id $pid_ -Force
        Write-Host " [stopped] e-cam-web vite dev PID $pid_ on :5173" -ForegroundColor Green
      }
    } elseif ($pids.Count -gt 0) {
      Write-Host " [kept]    :5173 vite not started from $CamWebRepo - left alone" -ForegroundColor Yellow
    }
  }
}

if ($IncludeNginx) {
  $nginxExe = 'D:\Haven\nginx-1.25.4\nginx.exe'
  $nginxPrefix = 'D:\Haven\nginx-1.25.4\'
  if (Get-Process nginx -ErrorAction SilentlyContinue) {
    if ($PSCmdlet.ShouldProcess('nginx (graceful -s stop)', 'Stop')) {
      & $nginxExe -p $nginxPrefix -s stop
      Write-Host " [stopped] nginx (-p $nginxPrefix -s stop)" -ForegroundColor Green
    }
  } else { Write-Host 'nginx: not running' -ForegroundColor DarkGray }
} else {
  Write-Host 'nginx: not touched (managed by you; pass -IncludeNginx to stop gracefully)' -ForegroundColor DarkGray
}
