# Expose the Mini App through a public tunnel (CloudPub, Pinggy, localhost.run).
param(
  [ValidateSet("cloudpub", "pinggy", "localhost")]
  [string]$Provider = "cloudpub",
  [int]$Port = 3000,
  [switch]$SkipBot
)

$ErrorActionPreference = "Stop"

function Test-LocalPort {
  param([int]$TargetPort)
  try {
    $r = Invoke-WebRequest -Uri "http://127.0.0.1:$TargetPort" -UseBasicParsing -TimeoutSec 3
    return $r.StatusCode -ge 200
  } catch {
    return $false
  }
}

function Find-CloBinary {
  $cmd = Get-Command clo -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd.Source }

  $candidates = @(
    (Join-Path $PSScriptRoot "tools\clo.exe"),
    (Join-Path $PSScriptRoot "..\tools\clo.exe"),
    (Join-Path $env:LOCALAPPDATA "CloudPub\clo.exe"),
    (Join-Path $env:ProgramFiles "CloudPub\clo.exe")
  )
  foreach ($path in $candidates) {
    if (Test-Path $path) { return $path }
  }
  return $null
}

function Start-SshTunnelJob {
  param(
    [string]$Name,
    [string[]]$SshArgs
  )

  $logOut = Join-Path $env:TEMP "markettg-$Name.out.log"
  $logErr = Join-Path $env:TEMP "markettg-$Name.err.log"
  foreach ($path in @($logOut, $logErr)) {
    if (Test-Path $path) { Remove-Item $path -Force }
  }

  $argLine = ($SshArgs | ForEach-Object {
    if ($_ -match '\s') { '"' + $_ + '"' } else { $_ }
  }) -join " "

  $job = Start-Job -ScriptBlock {
    param($ArgsLine, $OutPath, $ErrPath)
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "ssh"
    $psi.Arguments = $ArgsLine
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.UseShellExecute = $false
    $psi.CreateNoWindow = $true
    $proc = [System.Diagnostics.Process]::Start($psi)
    $stdout = $proc.StandardOutput.ReadToEnd()
    $stderr = $proc.StandardError.ReadToEnd()
    $proc.WaitForExit()
    if ($stdout) { Set-Content -Path $OutPath -Value $stdout -Encoding utf8 }
    if ($stderr) { Set-Content -Path $ErrPath -Value $stderr -Encoding utf8 }
    return $proc.ExitCode
  } -ArgumentList $argLine, $logOut, $logErr

  return [pscustomobject]@{
    Job    = $job
    LogOut = $logOut
    LogErr = $logErr
  }
}

function Get-TunnelLogText {
  param([string]$OutPath, [string]$ErrPath)
  $parts = @()
  foreach ($path in @($OutPath, $ErrPath)) {
    if (Test-Path $path) {
      $parts += Get-Content $path -Raw -ErrorAction SilentlyContinue
    }
  }
  return ($parts -join "`n")
}

function Wait-TunnelUrl {
  param(
    [string[]]$Patterns,
    [object]$WaitTarget,
    [string]$OutPath,
    [string]$ErrPath,
    [int]$TimeoutSec = 90
  )

  $url = $null
  for ($i = 0; $i -lt $TimeoutSec; $i++) {
    Start-Sleep -Seconds 1
    $text = Get-TunnelLogText -OutPath $OutPath -ErrPath $ErrPath
    foreach ($pattern in $Patterns) {
      if ($text -match $pattern) {
        $url = $Matches[0]
        break
      }
    }
    if ($url) { break }

    if ($WaitTarget -is [System.Diagnostics.Process]) {
      if ($WaitTarget.HasExited) { break }
    } elseif ($WaitTarget -is [System.Management.Automation.Job]) {
      if ($WaitTarget.State -in @("Completed", "Failed", "Stopped")) { break }
    }
  }

  if (-not $url) {
    throw "Tunnel did not return a URL.`n$(Get-TunnelLogText -OutPath $OutPath -ErrPath $ErrPath)"
  }
  return $url
}

function Update-MiniAppUrl {
  param([string]$Url)

  $envFile = Join-Path $PSScriptRoot "..\.env"
  $lines = @()
  if (Test-Path $envFile) {
    $lines = Get-Content $envFile
    $lines = $lines | Where-Object { $_ -notmatch "^\s*MINI_APP_URL=" }
  }
  $lines += "MINI_APP_URL=$Url"
  Set-Content -Path $envFile -Value $lines -Encoding utf8
}

function Register-TelegramMenu {
  param([string]$Url)

  $envFile = Join-Path $PSScriptRoot "..\.env"
  $token = $env:TELEGRAM_BOT_TOKEN
  if (-not $token -and (Test-Path $envFile)) {
    $tokenLine = Select-String -Path $envFile -Pattern "^\s*TELEGRAM_BOT_TOKEN=(.+)$" | Select-Object -Last 1
    if ($tokenLine) { $token = $tokenLine.Matches[0].Groups[1].Value.Trim() }
  }
  if (-not $token) { return $null }

  $menu = "{`"menu_button`":{`"type`":`"web_app`",`"text`":`"Магазин`",`"web_app`":{`"url`":`"$Url`"}}}"
  Invoke-RestMethod -Method Post -Uri "https://api.telegram.org/bot$token/setChatMenuButton" -ContentType "application/json" -Body $menu | Out-Null
  $me = Invoke-RestMethod -Method Get -Uri "https://api.telegram.org/bot$token/getMe"
  if ($me.result.username) {
    return "https://t.me/$($me.result.username)"
  }
  return $null
}

if (-not (Test-LocalPort -TargetPort $Port)) {
  throw "Mini App is not running on http://127.0.0.1:$Port. Start it first: cd apps\mini-app; npm run dev"
}

$url = $null
$keepAlive = $null

switch ($Provider) {
  "cloudpub" {
    $clo = Find-CloBinary
    if (-not $clo) {
      throw @"
CloudPub client (clo.exe) not found.

1. Register at https://cloudpub.ru
2. Download clo for Windows and put clo.exe into scripts\tools\
3. Run: clo login
4. Run this script again

Or install CloudPub Desktop and publish HTTP port $Port from the GUI.
"@
    }

    $logOut = Join-Path $env:TEMP "markettg-cloudpub.out.log"
    $logErr = Join-Path $env:TEMP "markettg-cloudpub.err.log"
    foreach ($path in @($logOut, $logErr)) {
      if (Test-Path $path) { Remove-Item $path -Force }
    }

    $proc = Start-Process -FilePath $clo -ArgumentList @("publish", "http", "$Port") `
      -RedirectStandardOutput $logOut -RedirectStandardError $logErr -PassThru -NoNewWindow

    $url = Wait-TunnelUrl -Patterns @(
      "https://[a-zA-Z0-9.-]+\.cloudpub\.ru",
      "https://[a-zA-Z0-9.-]+\.cloudpub\.com"
    ) -WaitTarget $proc -OutPath $logOut -ErrPath $logErr
  }

  "pinggy" {
    $tunnel = Start-SshTunnelJob -Name "pinggy" -SshArgs @(
      "-p", "443",
      "-o", "ConnectTimeout=20",
      "-o", "StrictHostKeyChecking=accept-new",
      "-o", "ServerAliveInterval=30",
      "-R0:127.0.0.1:$Port",
      "qr@free.pinggy.io"
    )
    $keepAlive = $tunnel
    $url = Wait-TunnelUrl -Patterns @("https://[a-zA-Z0-9.-]+\.[a-zA-Z0-9.-]+") `
      -WaitTarget $tunnel.Job -OutPath $tunnel.LogOut -ErrPath $tunnel.LogErr
  }

  "localhost" {
    $tunnel = Start-SshTunnelJob -Name "localhost" -SshArgs @(
      "-o", "ConnectTimeout=20",
      "-o", "StrictHostKeyChecking=accept-new",
      "-R", "80:127.0.0.1:$Port",
      "nokey@localhost.run"
    )
    $keepAlive = $tunnel
    $url = Wait-TunnelUrl -Patterns @("https://[a-zA-Z0-9.-]+\.(localhost\.run|lhr\.life|lhr\.rocks)") `
      -WaitTarget $tunnel.Job -OutPath $tunnel.LogOut -ErrPath $tunnel.LogErr
  }
}

Update-MiniAppUrl -Url $url

$botLink = $null
if (-not $SkipBot) {
  $botLink = Register-TelegramMenu -Url $url
}

Write-Output $url
if ($botLink) { Write-Output $botLink }

if ($keepAlive) {
  Write-Host ""
  Write-Host "Tunnel is running in background. Keep this PowerShell session open or run the SSH command manually."
}
