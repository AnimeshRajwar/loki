param(
  [string]$Channel = "stable",
  [string]$Repo = "AnimeshRajwar/loki",
  [string]$InstallDir = "C:\Program Files\Loki"
)

function Ensure-Admin {
  $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
  if (-not $isAdmin) {
    Write-Host "Requesting elevation..."
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "powershell"
    $psi.Arguments = "-NoProfile -ExecutionPolicy Bypass -File `"$($MyInvocation.MyCommand.Path)`""
    $psi.Verb = "runas"
    try { [System.Diagnostics.Process]::Start($psi) | Out-Null; exit } catch { Write-Error "Elevation canceled"; exit 1 }
  }
}

Ensure-Admin

$arch = if ($env:PROCESSOR_ARCHITECTURE -match "ARM") { "arm64" } else { "amd64" }
$asset = "loki-windows-$arch.zip"
$url = "https://github.com/$Repo/releases/download/$Channel/$asset"

$tmp = Join-Path $env:TEMP ("loki-install-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
  $zip = Join-Path $tmp $asset
  Write-Host "Downloading $url"
  Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing

  Write-Host "Extracting"
  Expand-Archive -Path $zip -DestinationPath $tmp -Force

  $src = Get-ChildItem -Path $tmp -Recurse -Filter "loki.exe" | Select-Object -First 1
  if (-not $src) { Write-Error "loki.exe not found in archive"; exit 1 }

  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
  Copy-Item -Path $src.FullName -Destination (Join-Path $InstallDir "loki.exe") -Force

  # add to machine PATH if missing
  $machinePath = [Environment]::GetEnvironmentVariable("Path","Machine")
  if (-not ($machinePath.Split(';') -contains $InstallDir)) {
    $newPath = $machinePath.TrimEnd(';') + ";" + $InstallDir
    [Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")
    Write-Host "Added $InstallDir to machine PATH. Restart required for sessions to pick up."
  }

  Write-Host "Installed: $InstallDir\loki.exe"
  Write-Host "Run 'loki --help' in a new elevated session or after restart."
}
finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
