# LAN Chat 构建指南

## macOS 版本

macOS 版本已经构建完成，位于：
- `build/macos/LANChat-macOS.dmg` - DMG 安装包
- `build/macos/LAN Chat.app` - 应用程序包

### 使用方法
1. 双击 `LANChat-macOS.dmg` 挂载磁盘
2. 将 `LAN Chat.app` 拖到应用程序文件夹
3. 从启动台启动 `LAN Chat`

## Windows 版本

由于跨平台编译的复杂性，Windows 版本需要在 Windows 系统上构建。

### 前置要求
1. 安装 Go 1.18+：https://golang.org/dl/
2. 安装 Node.js 16+：https://nodejs.org/
3. 安装 Git：https://git-scm.com/

### 构建步骤

1. **复制源代码到 Windows 电脑**
   ```bash
   # 将整个项目目录复制到 Windows 电脑
   # 可以使用 U 盘、网络共享或 Git
   ```

2. **安装依赖**
   ```powershell
   # 在 PowerShell 中运行
   cd lanchat
   
   # 安装 Go 依赖
   go mod tidy
   
   # 安装前端依赖
   cd frontend
   npm install
   cd ..
   ```

3. **构建前端**
   ```powershell
   cd frontend
   npm run build
   cd ..
   ```

4. **构建 Windows 可执行文件**
   ```powershell
   # 构建可执行文件
   go build -tags desktop,production -o lanchat.exe .
   ```

5. **运行程序**
   ```powershell
   .\lanchat.exe
   ```

### 打包为安装程序（可选）

如果需要打包为安装程序，可以使用：
- Inno Setup：https://jrsoftware.org/isinfo.php
- NSIS：https://nsis.sourceforge.io/

## 测试

在三台设备上运行程序后：
1. 每台设备应该能在左侧看到其他两台设备
2. 显示在线设备数量（应该是 3）
3. 在聊天框输入消息，其他设备应该能收到

## 故障排除

### 设备无法发现
- 确保所有设备在同一局域网
- 检查防火墙设置，允许 UDP 5353 端口（mDNS）
- 检查防火墙设置，允许 TCP 9527 端口

### 消息无法发送
- 检查 TCP 9527 端口是否被防火墙阻止
- 查看控制台日志获取详细错误信息

## 配置文件

配置文件位置：
- macOS: `~/Library/Application Support/lanchat/config.json`
- Windows: `%APPDATA%\lanchat\config.json`

可以修改配置文件中的：
- `device_name` - 设备名称
- `username` - 用户名
- `tcp_listen_port` - TCP 监听端口（默认 9527）
- `download_dir` - 文件下载目录
