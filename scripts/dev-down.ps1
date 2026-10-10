#Requires -Version 5.1
<#
.SYNOPSIS
  Stop the services that dev-up.ps1 started, and ONLY those.

.DESCRIPTION
  For ports 9000 / 8001 / 8081, look at the listening process and kill it only
  when its executable lives under logs\dev\bin (i.e. a mono-built binary).
  Existing old-line instances (e.g. a legacy e-cam-service binary from another
  directory), the vite dev server and nginx are left alone unless explicitly
  requested via -IncludeCamWeb / -IncludeNginx.

  Restart pattern:  .\scripts\dev-up.ps1 -Restart

  NOTE: keep this file ASCII-only (see dev-up.ps1 header note).
#>
[CmdletBinding(SupportsShouldProcess)]
param(
  [string]$BinDir = '',
  [switch]$IncludeCamWeb,   # also stop the mono e-cam-web vite dev (:5173) if it is ours
  [switch]$IncludeNginx     # also stop nginx gracefully (-s stop); OFF by default
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$bin  = if ($BinDir) { $BinDir } else { Join-Path $root 'logs\dev\bin' }

$map = @{ 9000 = 'eiam.exe'; 8001 = 'e-cam-service.exe'; 8081 = 'opsagent-api.exe' }

function Get-ListenPids([int]$Port) {
  (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique)
}

foreach ($port in ($map.Keys | Sort-Object)) {
  $want = (Join-Path $bin $map[$port]).ToLowerInvariant()
  $pids = @(Get-ListenPids $port)
  if ($pids.Count -eq 0) { Write-Host "port ${port}: free (nothing to stop)" -ForegroundColor DarkGray; continue }
  foreach ($pid_ in $pids) {
    $proc = Get-Process -Id $pid_ -ErrorAction SilentlyContinue
    if (-not $proc) { Write-Host "port ${port}: PID $pid_ already gone" -ForegroundColor DarkGray; continue }
    $path = $proc.Path
    if ($path -and $path.ToLowerInvariant() -eq $want) {
      if ($PSCmdlet.ShouldProcess("$($proc.ProcessName) PID $pid_ on :$port", 'Stop mono-managed service')) {
        Stop-Process -Id $pid_ -Force
        Write-Host " [stopped] $($proc.ProcessName) PID $pid_ on :$port (mono build)" -ForegroundColor Green
      }
    } else {
      Write-Host " [kept]    port ${port} owned by $($proc.ProcessName) PID $pid_ (path: $path) - not a mono build, left alone" -ForegroundColor Yellow
    }
  }
}

if ($IncludeCamWeb) {
  $pids = @(Get-ListenPids 5173)
  foreach ($pid_ in $pids) {
    $cli = (Get-CimInstance Win32_Process -Filter "ProcessId=$pid_" -ErrorAction SilentlyContinue).CommandLine
    if ($cli -and $cli.ToLowerInvariant().Contains((Join-Path $root 'apps\e-cam-web').ToLowerInvariant())) {
      if ($PSCmdlet.ShouldProcess("vite dev PID $pid_ on :5173 (mono e-cam-web)", 'Stop')) {
        Stop-Process -Id $pid_ -Force
        Write-Host " [stopped] mono e-cam-web vite dev PID $pid_ on :5173" -ForegroundColor Green
      }
    } elseif ($pids.Count -gt 0) {
      Write-Host " [kept]    :5173 vite not started from this repo - left alone (add -IncludeCamWeb only affects mono instances)" -ForegroundColor Yellow
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