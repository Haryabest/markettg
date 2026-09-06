# Run Telegram bot on the host (Docker bridge cannot reach api.telegram.org on Windows).
param(
  [switch]$StopDockerBot = $true
)

$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent
$envFile = Join-Path $root ".env"
$botDir = Join-Path $root "bot"

if (-not (Test-Path $envFile)) {
  throw "Missing .env in repo root. Copy env.ex and fill TELEGRAM_BOT_TOKEN."
}

Get-Content $envFile | ForEach-Object {
  if ($_ -match '^\s*([^#=]+?)=(.*)$') {
    $name = $Matches[1].Trim()
    $value = $Matches[2].Trim()
    Set-Item -Path "Env:$name" -Value $value
  }
}

if (-not $env:TELEGRAM_BOT_TOKEN) {
  throw "TELEGRAM_BOT_TOKEN is empty in .env"
}

if ($StopDockerBot) {
  Push-Location $root
  try {
    docker compose stop bot 2>&1 | Out-Null
  } catch {
    # Bot container may already be stopped.
  } finally {
    Pop-Location
  }
}

$pidFile = Join-Path $env:TEMP "markettg-bot-local.pid"
if (Test-Path $pidFile) {
  $oldPid = Get-Content $pidFile -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($oldPid -match '^\d+$') {
    Stop-Process -Id ([int]$oldPid) -Force -ErrorAction SilentlyContinue
  }
}

$env:GATEWAY_URL = "http://localhost:8080"
$env:CATALOG_SERVICE_URL = "http://localhost:8082"

if (-not $env:MINI_APP_URL) {
  Write-Warning "MINI_APP_URL not set. Run scripts\open-miniap-lt.ps1 first for Telegram WebApp."
}

$python = "python"
if (Get-Command py -ErrorAction SilentlyContinue) {
  $null = py -3.12 -c "import sys" 2>$null
  if ($LASTEXITCODE -eq 0) { $python = "py -3.12" }
}

Push-Location $botDir
try {
  Invoke-Expression "$python -m pip install -q -r requirements.txt"
  Invoke-Expression "$python -m bot.main"
} finally {
  Pop-Location
}
