# LAN Chat - 局域网聊天与文件传输工具

一款跨平台的局域网即时通讯和文件传输桌面应用，无需服务器，开箱即用。支持 **macOS** 和 **Windows**。

![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## 功能特性

### 即时通讯
- **自动设备发现** — 基于 UDP 广播协议，同一局域网内设备自动发现，无需手动配置
- **群聊模式** — 消息实时广播给所有在线设备
- **在线状态** — 左侧实时显示在线设备列表，包含用户名、设备名、IP 地址和平台标识
- **心跳检测** — 自动检测设备上下线，离线设备灰显并定时清理

### 文件传输
- **发送文件** — 选择文件 → 选择接收设备 → 等待对方接受 → 自动开始传输
- **接收文件** — 收到文件请求弹窗，可选择接受或拒绝，支持自定义保存路径
- **实时进度** — 传输过程显示进度条、已传输大小、百分比
- **取消传输** — 发送方和接收方均可随时取消传输
- **MD5 校验** — 每个数据块进行 MD5 完整性校验，确保文件传输准确
- **文件消息** — 传输完成后在聊天记录中显示文件消息卡片，可一键打开文件所在位置

### 界面设计
- **扁平化现代 UI** — 简洁的聊天界面，类微信气泡消息样式
- **自适应布局** — 支持窗口缩放，设置最小尺寸约束
- **平台图标** — 自动识别设备平台（macOS / Windows / Linux）并显示对应头像
- **文件类型图标** — 根据文件扩展名自动匹配对应图标（文档、图片、视频、压缩包等）

## 技术栈

| 层级 | 技术 |
|------|------|
| **桌面框架** | [Wails v2](https://wails.io/) (Go + WebView) |
| **后端** | Go 1.18+ |
| **前端** | Vue 3 + Pinia + Tailwind CSS |
| **UI 组件** | Arco Design Vue |
| **构建工具** | Vite 3 |
| **设备发现** | UDP 广播 (port 9528) |
| **数据传输** | TCP + JSON 协议 (port 9527) |
| **文件协议** | 自定义二进制分块协议 (LNFT) |

## 快速开始

### macOS

1. 下载 `build/macos/LANChat-macOS.dmg`
2. 双击挂载 DMG 文件
3. 将 `LAN Chat.app` 拖到「应用程序」文件夹
4. 从启动台启动 `LAN Chat`

### Windows

1. 下载 `build/windows/lanchat.exe`
2. 双击运行即可（无需安装）

> **注意**：首次运行时 Windows Defender 可能会弹出安全提示，选择「仍要运行」即可。

## 使用方法

1. 在多台设备上启动 LAN Chat（需在同一局域网内）
2. 等待 5-10 秒，设备会自动发现彼此
3. 左侧显示在线设备列表
4. 在底部输入框输入消息，按 **Enter** 发送
5. 点击附件按钮 📎 可发送文件给指定设备

## 项目结构

```
lanchat/
├── main.go                      # 入口文件，Wails 窗口配置
├── internal/                    # Go 后端
│   ├── app/                     # Wails 绑定层 & 事件桥接
│   │   ├── app.go               #   应用主逻辑，暴露前端方法
│   │   ├── eventbridge.go       #   Wails 事件发射器
│   │   ├── dialog_darwin.go     #   macOS 文件选择对话框
│   │   ├── dialog_windows.go    #   Windows 文件选择对话框 (Win32 API)
│   │   └── tray_other.go        #   系统托盘（占位）
│   ├── chat/                    # 聊天服务
│   │   └── service.go           #   消息发送、接收、存储
│   ├── config/                  # 配置管理
│   │   └── config.go            #   加载/保存/创建默认配置
│   ├── discovery/               # 设备发现
│   │   └── discovery_udp.go     #   UDP 广播发现协议
│   ├── filetransfer/            # 文件传输
│   │   ├── manager.go           #   传输状态管理、请求/接受/拒绝/取消
│   │   └── transfer.go          #   分块发送器 & 接收器 (LNFT 协议)
│   ├── models/                  # 数据模型
│   │   ├── device.go            #   设备结构体
│   │   ├── message.go           #   聊天消息结构体
│   │   └── transfer.go          #   文件传输结构体
│   ├── peer/                    # 设备管理
│   │   └── manager.go           #   设备增删、心跳、超时检测
│   └── transport/               # TCP 传输层
│       ├── client.go            #   连接 & 发送
│       ├── codec.go             #   消息编解码 (Header + JSON)
│       └── server.go            #   TCP 监听 & 消息路由
├── frontend/                    # Vue 3 前端
│   ├── src/
│   │   ├── App.vue              #   根组件，事件监听
│   │   ├── main.js              #   入口，开发环境加载 Mock
│   │   ├── style.css            #   全局样式 (Tailwind)
│   │   ├── components/
│   │   │   ├── ChatView.vue     #     聊天视图 & 消息输入
│   │   │   ├── DeviceList.vue   #     在线设备列表
│   │   │   ├── FileRequestDialog.vue  # 文件接收请求弹窗
│   │   │   ├── FileReceivePanel.vue   # 文件接收进度面板
│   │   │   └── FileTransferPanel.vue  # 文件发送面板
│   │   ├── stores/
│   │   │   ├── chat.js          #     聊天状态 (Pinia)
│   │   │   └── devices.js       #     设备状态 (Pinia)
│   │   └── utils/
│   │       └── file.js          #     文件图标映射 & 大小格式化
│   ├── wailsjs/                 #     Wails 自动生成的绑定
│   └── ...
├── build/                       # 构建资源 & 产物
│   ├── appicon.png              #   应用图标源文件
│   ├── darwin/                  #   macOS 构建配置
│   ├── macos/                   #   macOS 构建产物
│   └── windows/                 #   Windows 构建配置 & 产物
├── build-windows.ps1            # Windows 构建脚本
└── wails.json                   # Wails 项目配置
```

## 开发

### 前置要求

- Go 1.18+
- Node.js 16+
- npm 或 yarn

### 安装依赖

```bash
# Go 依赖
go mod tidy

# 前端依赖
cd frontend && npm install && cd ..
```

### 开发模式

```bash
# 启动前端开发服务器
cd frontend && npm run dev

# 另一个终端，启动后端
go run -tags desktop,production .
```

### 构建生产版本

**macOS：**
```bash
# 构建前端
cd frontend && npm run build && cd ..

# 构建 macOS 二进制
CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags desktop,production -o lanchat .

# 生成 DMG 安装包
mkdir -p build/macos/dmg_staging
cp -R "build/macos/LAN Chat.app" build/macos/dmg_staging/
ln -s /Applications build/macos/dmg_staging/Applications
hdiutil create -volname "LAN Chat" -srcfolder build/macos/dmg_staging \
  -ov -format UDZO build/macos/LANChat-macOS.dmg
```

**Windows（在 macOS 上交叉编译）：**
```bash
# 安装交叉编译工具链
brew install mingw-w64

# 构建
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -tags desktop,production -ldflags="-H=windowsgui" -o lanchat.exe .
```

**Windows（在 Windows 上构建）：**
```powershell
.\build-windows.ps1
```

## 配置文件

配置文件位置：
- **macOS**: `~/Library/Application Support/lanchat/config.json`
- **Windows**: `%APPDATA%\lanchat\config.json`

```json
{
  "device_id": "自动生成的 UUID",
  "device_name": "设备名称（默认为主机名）",
  "username": "用户名",
  "download_dir": "文件下载目录",
  "tcp_listen_port": 9527
}
```

## 网络要求

| 端口 | 协议 | 用途 |
|------|------|------|
| 9527 | TCP | 聊天消息 & 文件传输控制信令 |
| 9528 | UDP | 设备广播发现 |
| 随机端口 | TCP | 文件数据传输通道 |

> 请确保局域网防火墙允许上述端口通信。

## 故障排除

**设备无法发现**
- 确认所有设备在同一局域网 / 子网
- 检查防火墙是否放行 UDP 9528 端口

**消息无法送达**
- 检查防火墙是否放行 TCP 9527 端口
- 查看终端日志获取详细错误信息

**文件传输失败**
- 确认两端网络连通
- 文件传输使用随机 TCP 端口，确保防火墙未阻止

## 许可证

MIT License
