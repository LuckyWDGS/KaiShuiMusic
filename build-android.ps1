# 构建开水音乐安卓 APP
# 产物：dist\KaiShuiMusic-<版本>.apk
# 依赖：Node.js 18+、Go 1.25+、JDK 17、Android SDK（ANDROID_HOME）和 NDK、gomobile
#   go install golang.org/x/mobile/cmd/gomobile@latest; go install golang.org/x/mobile/cmd/gobind@latest
# 发布签名：android\keystore.properties（storeFile、storePassword、keyAlias、keyPassword），没有时使用调试签名
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

foreach ($dir in '..\_dependencies\go\bin', '..\_dependencies\gobin') {
    $p = Join-Path $PSScriptRoot $dir
    if (Test-Path $p) { $env:PATH = "$p;$env:PATH" }
}
if (-not $env:ANDROID_HOME) { $env:ANDROID_HOME = Join-Path $env:LOCALAPPDATA 'Android\Sdk' }
if (-not $env:ANDROID_NDK_HOME) {
    $ndks = Get-ChildItem (Join-Path $env:ANDROID_HOME 'ndk') | Sort-Object Name
    $env:ANDROID_NDK_HOME = ($ndks | Where-Object Name -Like "28.*" | Select-Object -Last 1).FullName
    if (-not $env:ANDROID_NDK_HOME) { $env:ANDROID_NDK_HOME = $ndks[-1].FullName }
}
# 中文 Windows 下 javac 默认 GBK，读不了 gomobile 生成的 UTF-8 源码
$env:JAVA_TOOL_OPTIONS = '-Dfile.encoding=UTF-8'
$version = (Get-Content VERSION -Raw).Trim()

Write-Host '==> 构建前端' -ForegroundColor Cyan
Push-Location web
if (-not (Test-Path node_modules)) { npm install }
npm run build
Pop-Location

Write-Host '==> 构建 Go 内核' -ForegroundColor Cyan
New-Item -ItemType Directory -Force android\app\libs | Out-Null
Push-Location server
gomobile bind -target='android/arm64,android/arm,android/amd64' -androidapi 24 -javapkg com.kaishui.go -ldflags '-s -w' -o ..\android\app\libs\kaishui.aar .\mobile
if ($LASTEXITCODE) { throw 'gomobile bind 失败' }
Pop-Location

Write-Host '==> 打包 APK' -ForegroundColor Cyan
Push-Location android
.\gradlew.bat assembleRelease --offline
if ($LASTEXITCODE) { .\gradlew.bat assembleRelease }
if ($LASTEXITCODE) { throw 'APK 打包失败' }
Pop-Location

New-Item -ItemType Directory -Force dist | Out-Null
Copy-Item android\app\build\outputs\apk\release\app-release.apk "dist\KaiShuiMusic-$version.apk" -Force
Write-Host "==> 完成：dist\KaiShuiMusic-$version.apk" -ForegroundColor Green
