package app

import (
	"context"
	"fmt"
	"lanchat/internal/chat"
	"lanchat/internal/config"
	"lanchat/internal/discovery"
	"lanchat/internal/filetransfer"
	"lanchat/internal/models"
	"lanchat/internal/peer"
	"lanchat/internal/transport"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type App struct {
	ctx       context.Context
	cfg       *config.Config
	discovery *discovery.Discovery
	peers     *peer.Manager
	chat      *chat.Service
	fileMgr   *filetransfer.Manager
	server    *transport.Server
	events    *EventEmitter
	cancel    context.CancelFunc
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	a.cfg = cfg

	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	a.events = NewEventEmitter(ctx)

	a.peers = peer.NewManager(a.events)

	a.discovery = discovery.New(cfg)
	if err := a.discovery.Register(ctx); err != nil {
		log.Printf("Failed to register UDP broadcast: %v", err)
	}

	a.chat = chat.NewService(cfg.DeviceID, cfg.DeviceName, a.peers, a.events)

	a.fileMgr = filetransfer.NewManager(a.events, a.peers, cfg.DeviceID, cfg.DeviceName)
	a.fileMgr.SetChatNotifier(a.chat)

	a.server = transport.NewServer(cfg.TCPListenPort, func(conn net.Conn, msg *transport.Message) {
		// 根据消息类型路由到不同的处理器
		switch msg.Type {
		case "file_request":
			a.fileMgr.HandleFileRequest(msg.Data)
		case "file_accept":
			a.fileMgr.HandleFileAccept(msg.Data)
		case "file_reject":
			a.fileMgr.HandleFileReject(msg.Data)
		case "file_cancel":
			a.fileMgr.HandleFileCancel(msg.Data)
		default:
			a.chat.HandleMessage(conn, msg)
		}
	})
	go func() {
		if err := a.server.Start(ctx); err != nil {
			log.Printf("TCP server error: %v", err)
		}
	}()

	go func() {
		if err := a.discovery.Browse(ctx, a.peers.AddOrUpdate); err != nil {
			log.Printf("UDP browse error: %v", err)
		}
	}()

	go a.peers.StartHeartbeat(ctx)

	log.Printf("App started: %s@%s (ID: %s)", cfg.Username, cfg.DeviceName, cfg.DeviceID)
}

func (a *App) Shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	log.Println("App shutdown")
}

func (a *App) GetDevices() []*models.Device {
	return a.peers.GetAll()
}

func (a *App) GetOnlineDevices() []*models.Device {
	return a.peers.GetOnline()
}

func (a *App) GetOnlineCount() int {
	return a.peers.GetOnlineCount()
}

func (a *App) SendMessage(content string) error {
	return a.chat.Send(a.ctx, content)
}

func (a *App) GetMessages() []*models.ChatMessage {
	return a.chat.GetMessages()
}

func (a *App) GetConfig() *config.Config {
	return a.cfg
}

func (a *App) AddPeerManually(ip string, port int) error {
	log.Printf("[App] Manually adding peer: %s:%d", ip, port)
	
	device := &models.Device{
		DeviceID:   "manual-" + ip,
		DeviceName: "Manual Peer",
		Username:   "user",
		IP:         ip,
		Port:       port,
		Platform:   "unknown",
		Status:     models.StatusOnline,
	}
	
	a.peers.AddOrUpdate(device)
	return nil
}

// 文件传输相关方法
func (a *App) SendFile(peerID, filePath string) error {
	return a.fileMgr.SendFile(a.ctx, peerID, filePath, a.cfg.DeviceID)
}

func (a *App) AcceptFile(transferID, savePath string) error {
	return a.fileMgr.AcceptFile(transferID, savePath)
}

func (a *App) RejectFile(transferID string) error {
	return a.fileMgr.RejectFile(transferID)
}

func (a *App) CancelFile(transferID string) error {
	return a.fileMgr.CancelFile(transferID)
}

func (a *App) GetTransfers() []*models.FileTransfer {
	return a.fileMgr.GetAllTransfers()
}

// SelectFile 打开文件选择对话框
func (a *App) SelectFile() (string, error) {
	log.Printf("[App] SelectFile called")
	filePath, err := openFileDialog(a.ctx)
	if err != nil {
		log.Printf("[App] SelectFile error: %v", err)
		return "", err
	}
	log.Printf("[App] SelectFile result: %s", filePath)
	return filePath, nil
}

// SelectSavePath 打开保存文件对话框
func (a *App) SelectSavePath(fileName string) (string, error) {
	log.Printf("[App] SelectSavePath called: %s", fileName)

	// 使用用户下载目录作为默认保存路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	downloadDir := filepath.Join(homeDir, "Downloads")

	// 确保目录存在
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("create download dir: %w", err)
	}

	savePath := filepath.Join(downloadDir, fileName)
	log.Printf("[App] SelectSavePath default path: %s", savePath)
	return savePath, nil
}

// OpenFileLocation 打开文件所在目录
func (a *App) OpenFileLocation(filePath string) error {
	log.Printf("[App] OpenFileLocation: %s", filePath)

	dir := filepath.Dir(filePath)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-R", filePath)
	case "windows":
		cmd = exec.Command("explorer", "/select,", filePath)
	default:
		cmd = exec.Command("xdg-open", dir)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open file location: %w", err)
	}
	return nil
}

type FileInfoResult struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func (a *App) GetFileInfo(filePath string) (*FileInfoResult, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	return &FileInfoResult{
		Name: filepath.Base(filePath),
		Size: stat.Size(),
	}, nil
}
