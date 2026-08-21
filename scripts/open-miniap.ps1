# Expose the Mini App through fxTunnel and register the HTTPS URL with Telegram.
param(
  [int]$Port = 3000,
  [string]$Domain = "markettg"
)

$ErrorActionPreference = "Stop"
$fxtunnel = Join-Path $env:LOCALAPPDATA "fxTunnel\fxtunnel.exe"
if (-not (Test-Path $fxtunnel)) {
  throw "fxtunnel.exe not found at $fxtunnel"
}

$logOut = Join-Path $env:TEMP "markettg-fxtunnel.out.log"
$logErr = Join-Path $env:TEMP "markettg-fxtunnel.err.log"
foreach ($path in @($logOut, $logErr)) {
  if (Test-Path $path) { Remove-Item $path -Force }
}

function Get-FxTunnelLogText {
  $parts = @()
  foreach ($path in @($logOut, $logErr)) {
    if (Test-Path $path) {
      $parts += Get-Content $path -Raw -ErrorAction SilentlyContinue
    }
  }
  return ($parts -join "`n")
}

$tunnel = Start-Process -FilePath $fxtunnel -ArgumentList @("http", "$Port", "--domain", $Domain) `
  -RedirectStandardOutput $logOut -RedirectStandardError $logErr -PassThru -NoNewWindow
$url = $null
for ($i = 0; $i -lt 120; $i++) {
  Start-Sleep -Seconds 1
  $text = Get-FxTunnelLogText
  if ($text -match "https://[a-zA-Z0-9.-]+\.fxtun\.dev") {
    $url = $Matches[0]
    break
  }
  if ($tunnel.HasExited) { break }
}

if (-not $url) {
  throw "fxTunnel did not return a URL.`n$(Get-FxTunnelLogText)"
}

$envFile = Join-Path $PSScriptRoot "..\.env"
$lines = @()
if (Test-Path $envFile) {
  $lines = Get-Content $envFile
  $lines = $lines | Where-Object { $_ -notmatch "^\s*MINI_APP_URL=" }
}
$lines += "MINI_APP_URL=$url"
Set-Content -Path $envFile -Value $lines -Encoding utf8

$token = $env:TELEGRAM_BOT_TOKEN
if (-not $token -and (Test-Path $envFile)) {
  $tokenLine = Select-String -Path $envFile -Pattern "^\s*TELEGRAM_BOT_TOKEN=(.+)$" | Select-Object -Last 1
  if ($tokenLine) { $token = $tokenLine.Matches[0].Groups[1].Value.Trim() }
}

$botLink = $null
if ($token) {
  $encoded = [uri]::EscapeDataString($url)
  $menu = "{`"menu_button`":{`"type`":`"web_app`",`"text`":`"Магазин`",`"web_app`":{`"url`":`"$url`"}}}"
  Invoke-RestMethod -Method Post -Uri "https://api.telegram.org/bot$token/setChatMenuButton" -ContentType "application/json" -Body $menu | Out-Null
  $me = Invoke-RestMethod -Method Get -Uri "https://api.telegram.org/bot$token/getMe"
  if ($me.result.username) {
    $botLink = "https://t.me/$($me.result.username)"
  }
}

Write-Output $url
if ($botLink) { Write-Output $botLink }
