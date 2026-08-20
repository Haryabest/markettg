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

$log = Join-Path $env:TEMP "markettg-fxtunnel.log"
if (Test-Path $log) { Remove-Item $log -Force }

$tunnel = Start-Process -FilePath $fxtunnel -ArgumentList @("http", "$Port", "--domain", $Domain) -RedirectStandardOutput $log -RedirectStandardError $log -PassThru -NoNewWindow
$url = $null
for ($i = 0; $i -lt 40; $i++) {
  Start-Sleep -Seconds 1
  if (Test-Path $log) {
    $text = Get-Content $log -Raw -ErrorAction SilentlyContinue
    if ($text -match "https://[a-zA-Z0-9.-]+\.fxtun\.dev") {
      $url = $Matches[0]
      break
    }
  }
  if ($tunnel.HasExited) { break }
}

if (-not $url) {
  $fallback = Get-Content $log -Raw -ErrorAction SilentlyContinue
  throw "fxTunnel did not return a URL.`n$fallback"
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
