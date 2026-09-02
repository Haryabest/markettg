# Expose Mini App + API via localtunnel (port 80 = Caddy: mini-app + /api/*).
param(
  [int]$Port = 80,
  [string]$Subdomain = "markettg674598",
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
  $null = Invoke-WebRequest -Uri "http://127.0.0.1:$Port" -UseBasicParsing -TimeoutSec 5
} catch {
  throw "Nothing is listening on http://127.0.0.1:$Port. Start Docker: docker compose up -d caddy gateway mini-app"
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

$ltArgs = @("--yes", "localtunnel", "--port", $Port)
if ($Subdomain) {
  $ltArgs += @("--subdomain", $Subdomain)
}

$command = "npx $($ltArgs -join ' ') 2>&1 | Tee-Object -FilePath '$log' -Append"
$proc = Start-Process -FilePath "powershell.exe" `
  -ArgumentList @("-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", $command) `
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
  $lines = $lines | Where-Object {
    $_ -notmatch "^\s*MINI_APP_URL=" -and $_ -notmatch "^\s*NEXT_PUBLIC_API_URL="
  }
}
$lines += "MINI_APP_URL=$url"
$lines += "NEXT_PUBLIC_API_URL=$url"
Set-Content -Path $envFile -Value $lines -Encoding utf8

if (-not $SkipBot) {
  $token = $env:TELEGRAM_BOT_TOKEN
  if (-not $token -and (Test-Path $envFile)) {
    $tokenLine = Select-String -Path $envFile -Pattern "^\s*TELEGRAM_BOT_TOKEN=(.+)$" | Select-Object -Last 1
    if ($tokenLine) { $token = $tokenLine.Matches[0].Groups[1].Value.Trim() }
  }
  if ($token) {
    try {
      $menu = '{"menu_button":{"type":"web_app","text":"Shop","web_app":{"url":"' + $url + '"}}}'
      Invoke-RestMethod -Method Post -Uri "https://api.telegram.org/bot$token/setChatMenuButton" -ContentType "application/json; charset=utf-8" -Body ([System.Text.Encoding]::UTF8.GetBytes($menu)) | Out-Null
      $cmdsObj = @{
        commands = @(
          @{ command = "start"; description = "Main menu" },
          @{ command = "catalog"; description = "Catalog" },
          @{ command = "orders"; description = "My orders" },
          @{ command = "support"; description = "Support" },
          @{ command = "legal"; description = "Legal info" },
          @{ command = "help"; description = "Help" }
        )
      }
      $cmds = $cmdsObj | ConvertTo-Json -Depth 5 -Compress
      Invoke-RestMethod -Method Post -Uri "https://api.telegram.org/bot$token/setMyCommands" -ContentType "application/json; charset=utf-8" -Body ([System.Text.Encoding]::UTF8.GetBytes($cmds)) | Out-Null
      Write-Host "Telegram menu button and commands updated."
    } catch {
      Write-Warning "Could not update bot via Telegram API: $($_.Exception.Message)"
    }
  }
}

Write-Output $url
Write-Host ""
Write-Host "Tunnel PID wrapper: $($proc.Id). Keep it running while testing in Telegram."
Write-Host "First visit may show loca.lt warning page - click Continue."
