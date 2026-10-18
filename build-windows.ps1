# LAN Chat Windows 构建脚本
# 在 PowerShell 中运行此脚本

Write-Host "=== LAN Chat Windows 构建脚本 ===" -ForegroundColor Green

# 检查 Go 是否安装
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "错误: 未找到 Go，请先安装 Go 1.18+" -ForegroundColor Red
    Write-Host "下载地址: https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}

# 检查 Node.js 是否安装
if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Host "错误: 未找到 Node.js，请先安装 Node.js 16+" -ForegroundColor Red
    Write-Host "下载地址: https://nodejs.org/" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ Go 版本: $(go version)" -ForegroundColor Green
Write-Host "✓ Node.js 版本: $(node --version)" -ForegroundColor Green

# 安装 Go 依赖
Write-Host "`n正在安装 Go 依赖..." -ForegroundColor Cyan
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "错误: Go 依赖安装失败" -ForegroundColor Red
    exit 1
}

# 安装前端依赖
Write-Host "`n正在安装前端依赖..." -ForegroundColor Cyan
Set-Location frontend
npm install
if ($LASTEXITCODE -ne 0) {
    Write-Host "错误: 前端依赖安装失败" -ForegroundColor Red
    exit 1
}

# 构建前端
Write-Host "`n正在构建前端..." -ForegroundColor Cyan
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Host "错误: 前端构建失败" -ForegroundColor Red
    exit 1
}
Set-Location ..

# 构建 Windows 可执行文件
Write-Host "`n正在构建 Windows 可执行文件..." -ForegroundColor Cyan
go build -tags desktop,production -ldflags="-H=windowsgui" -o lanchat.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "错误: 构建失败" -ForegroundColor Red
    exit 1
}

Write-Host "`n✓ 构建成功!" -ForegroundColor Green
Write-Host "可执行文件: lanchat.exe" -ForegroundColor Green
Write-Host "`n运行程序: .\lanchat.exe" -ForegroundColor Cyan
