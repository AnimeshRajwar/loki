param(
  [ValidateSet("stable", "latest")]
  [string]$Channel = "stable",
  [string]$Repo = "AnimeshRajwar/loki",
  [string]$InstallDir = "$env:LOCALAPPDATA\Programs\Loki"
)

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$asset = "loki-windows-$arch.zip"
$url = "https://github.com/$Repo/releases/download/$Channel/$asset"

$tmp = Join-Path $env:TEMP ("loki-install-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
  $zip = Join-Path $tmp $asset
  Invoke-WebRequest -Uri $url -OutFile $zip
  Expand-Archive -Path $zip -DestinationPath $tmp -Force

  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
  $sourceExe = Join-Path $tmp "loki-windows-$arch\loki.exe"
  $targetExe = Join-Path $InstallDir "loki.exe"
  Copy-Item -Path $sourceExe -Destination $targetExe -Force

  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  if ([string]::IsNullOrWhiteSpace($userPath)) {
    [Environment]::SetEnvironmentVariable("Path", $InstallDir, "User")
    Write-Host "Created user PATH with: $InstallDir"
  }
  elseif (-not ($userPath.Split(';') -contains $InstallDir)) {
    [Environment]::SetEnvironmentVariable("Path", ($userPath.TrimEnd(';') + ";$InstallDir"), "User")
    Write-Host "Added to user PATH: $InstallDir"
  }

  Write-Host "Installed: $targetExe"
  Write-Host "Open a new terminal and run: loki --help"
}
finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
