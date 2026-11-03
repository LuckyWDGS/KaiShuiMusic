# 构建开水音乐 PC 版（Windows）
# 产物：dist\KaiShuiMusic.exe（桌面版，双击运行）、dist\kaishui.exe（服务版，浏览器 / 手机 / 电视访问）
# 依赖：Node.js 18+、Go 1.25+（未加入 PATH 时自动尝试 ..\_dependencies\go）
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    $localGo = Join-Path $PSScriptRoot '..\_dependencies\go\bin'
    if (Test-Path $localGo) { $env:PATH = "$localGo;$env:PATH" } else { throw '未找到 Go，请先安装 https://go.dev/dl/' }
}

Write-Host '==> 构建前端' -ForegroundColor Cyan
Push-Location web
if (-not (Test-Path node_modules)) { npm install }
npm run build
Pop-Location

Write-Host '==> 构建后端' -ForegroundColor Cyan
Push-Location server
# 修改了 winres/ 下的图标或清单时重新生成资源（需要 go install github.com/tc-hib/go-winres@latest）
$winres = Get-Command go-winres -ErrorAction SilentlyContinue
if (-not $winres) { $winres = Get-Command (Join-Path $PSScriptRoot '..\_dependencies\gobin\go-winres.exe') -ErrorAction SilentlyContinue }
if ($winres) { & $winres.Source make --in winres\winres.json --out cmd\kaishui-desktop\rsrc }
go build -trimpath -ldflags '-s -w -H windowsgui' -o ..\dist\KaiShuiMusic.exe .\cmd\kaishui-desktop
go build -trimpath -ldflags '-s -w' -o ..\dist\kaishui.exe .\cmd\kaishui
Pop-Location

Write-Host '==> 完成：dist\KaiShuiMusic.exe' -ForegroundColor Green
