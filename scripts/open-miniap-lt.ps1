# Expose Mini App via localtunnel (works when fxtunnel/Pinggy are blocked).
param(
  [int]$Port = 3000,
  [switch]$SkipBot
)

$ErrorActionPreference = "Stop"

function Stop-PreviousTunnel {
  $pidFile = Join-Path $env:TEMP "markettg-localtunnel.pid"
  if (-not (Test-Path $pidFile)) { return }

  $oldPid = Get-Content $pidFile -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($oldPid -match '^\d+$') {
    Stop-Process -Id ([int]$oldPid) -Force -ErrorAction SilentlyContinue
    Start-Sleep -Milliseconds 500
  }
  Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
}

try {
  $null = Invoke-WebRequest -Uri "http://127.0.0.1:$Port" -UseBasicParsing -TimeoutSec 3
} catch {
  throw "Mini App is not running on http://127.0.0.1:$Port. Start: cd apps\mini-app; npm run dev"
}

Stop-PreviousTunnel

$log = Join-Path $env:TEMP "markettg-localtunnel.log"
if (Test-Path $log) {
  try {
    Remove-Item $log -Force -ErrorAction Stop
  } catch {
    $log = Join-Path $env:TEMP ("markettg-localtunnel-{0}.log" -f [guid]::NewGuid().ToString("N"))
  }
}

$miniAppDir = Join-Path $PSScriptRoot "..\apps\mini-app"
$command = "npx --yes localtunnel --port $Port 2>&1 | Tee-Object -FilePath '$log' -Append"
$proc = Start-Process -FilePath "powershell.exe" `
  -ArgumentList @("-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "Set-Location '$miniAppDir'; $command") `
  -PassThru -WindowStyle Hidden

Set-Content -Path (Join-Path $env:TEMP "markettg-localtunnel.pid") -Value $proc.Id -Encoding ascii

$url = $null
for ($i = 0; $i -lt 90; $i++) {
  Start-Sleep -Seconds 1
  if (-not (Test-Path $log)) { continue }
  $text = Get-Content $log -Raw -ErrorAction SilentlyContinue
  $matches = [regex]::Matches($text, "https://[a-zA-Z0-9-]+\.loca\.lt")
  if ($matches.Count -gt 0) {
    $url = $matches[$matches.Count - 1].Value
    break
  }
  if ($proc.HasExited) { break }
}

if (-not $url) {
  throw "localtunnel did not return a URL.`n$(Get-Content $log -Raw -ErrorAction SilentlyContinue)"
}

$envFile = Join-Path $PSScriptRoot "..\.env"
$lines = @()
if (Test-Path $envFile) {
  $lines = Get-Content $envFile
  $lines = $lines | Where-Object { $_ -notmatch "^\s*MINI_APP_URL=" }
}
$lines += "MINI_APP_URL=$url"
Set-Content -Path $envFile -Value $lines -Encoding utf8

if (-not $SkipBot) {
  $token = $env:TELEGRAM_BOT_TOKEN
  if (-not $token -and (Test-Path $envFile)) {
    $tokenLine = Select-String -Path $envFile -Pattern "^\s*TELEGRAM_BOT_TOKEN=(.+)$" | Select-Object -Last 1
    if ($tokenLine) { $token = $tokenLine.Matches[0].Groups[1].Value.Trim() }
  }
  if ($token) {
    $menu = "{`"menu_button`":{`"type`":`"web_app`",`"text`":`"Магазин`",`"web_app`":{`"url`":`"$url`"}}}"
    Invoke-RestMethod -Method Post -Uri "https://api.telegram.org/bot$token/setChatMenuButton" -ContentType "application/json" -Body $menu | Out-Null
  }
}

Write-Output $url
Write-Host ""
Write-Host "Tunnel PID wrapper: $($proc.Id). Keep it running while the client tests."
Write-Host "First visit may show loca.lt warning page - click Continue."
