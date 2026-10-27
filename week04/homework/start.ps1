$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$serverDir = Join-Path $projectRoot 'server'
$clientDir = Join-Path $projectRoot 'client'

Write-Host '正在启动 Go 后端：http://localhost:8080' -ForegroundColor Cyan
Start-Process powershell -WindowStyle Normal -ArgumentList @(
    '-NoExit',
    '-ExecutionPolicy', 'Bypass',
    '-Command', "Set-Location -LiteralPath '$serverDir'; go run ."
)

Write-Host '正在启动 React 前端：http://localhost:5173' -ForegroundColor Cyan
Start-Process powershell -WindowStyle Normal -ArgumentList @(
    '-NoExit',
    '-ExecutionPolicy', 'Bypass',
    '-Command', "Set-Location -LiteralPath '$clientDir'; npm run dev"
)

Write-Host '首次连接云数据库可能需要几秒。服务启动后请打开：http://localhost:5173' -ForegroundColor Green
