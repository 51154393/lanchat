package filetransfer

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"lanchat/internal/models"
	"lanchat/internal/transport"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type EventEmitter interface {
	Emit(eventName string, data interface{})
}

type PeerGetter interface {
	Get(deviceID string) *models.Device
}

type ChatNotifier interface {
	AddFileMessage(fileName, filePath string, fileSize int64, senderID, senderName string, isSender bool)
}

type Manager struct {
	transfers map[string]*models.FileTransfer
	dataConns map[string]net.Conn
	mu        sync.RWMutex
	events    EventEmitter
	peers     PeerGetter
	chat      ChatNotifier
	deviceID  string
	deviceName string
}

func NewManager(events EventEmitter, peers PeerGetter, deviceID, deviceName string) *Manager {
	return &Manager{
		transfers:  make(map[string]*models.FileTransfer),
		dataConns:  make(map[string]net.Conn),
		events:     events,
		peers:      peers,
		deviceID:   deviceID,
		deviceName: deviceName,
	}
}

func (m *Manager) SetChatNotifier(chat ChatNotifier) {
	m.chat = chat
}

// SendFile 发送文件请求
func (m *Manager) SendFile(ctx context.Context, peerID, filePath string, senderID string) error {
	peer := m.peers.Get(peerID)
	if peer == nil {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	// 获取文件信息
	fileInfo, err := getFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("get file info: %w", err)
	}

	transferID := uuid.New().String()
	transfer := &models.FileTransfer{
		TransferID:  transferID,
		PeerID:      peerID,
		Direction:   models.TransferSending,
		Status:      models.TransferPending,
		FileName:    fileInfo.Name,
		FileSize:    fileInfo.Size,
		FileHash:    fileInfo.Hash,
		LocalPath:   filePath,
		ChunkSize:   65536, // 64KB
		StartedAt:   time.Now(),
	}

	m.mu.Lock()
	m.transfers[transferID] = transfer
	m.mu.Unlock()

	// 发送文件请求
	reqData := FileRequestData{
		TransferID: transferID,
		SenderID:   senderID,
		FileName:   fileInfo.Name,
		FileSize:   fileInfo.Size,
		FileHash:   fileInfo.Hash,
		MimeType:   fileInfo.MimeType,
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	conn, err := transport.Dial(peer.IP, peer.Port)
	if err != nil {
		return fmt.Errorf("connect to peer: %w", err)
	}
	defer conn.Close()

	msg := &transport.Message{
		Type: "file_request",
		Data: data,
	}

	if err := transport.Send(conn, msg); err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	log.Printf("[FileTransfer] Sent file request: %s to %s", fileInfo.Name, peer.DeviceName)

	if m.events != nil {
		m.events.Emit("file:transfer_started", transfer)
	}

	return nil
}

// AcceptFile 接受文件传输
func (m *Manager) AcceptFile(transferID string, savePath string) error {
	log.Printf("[FileTransfer] AcceptFile called: transferID=%s, savePath=%s", transferID, savePath)

	m.mu.Lock()
	transfer, exists := m.transfers[transferID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("transfer not found: %s", transferID)
	}
	transfer.Status = models.TransferAccepted
	transfer.LocalPath = savePath
	m.mu.Unlock()

	log.Printf("[FileTransfer] Accepted file: %s, save to: %s", transfer.FileName, savePath)

	// receiveFile 会负责监听端口、发送 file_accept（带 DataPort）并开始接收
	go m.receiveFile(transfer)

	log.Printf("[FileTransfer] receiveFile goroutine started")
	return nil
}

// RejectFile 拒绝文件传输
func (m *Manager) RejectFile(transferID string) error {
	m.mu.Lock()
	transfer, exists := m.transfers[transferID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("transfer not found: %s", transferID)
	}
	transfer.Status = models.TransferRejected
	m.mu.Unlock()

	// 发送拒绝响应
	peer := m.peers.Get(transfer.PeerID)
	if peer == nil {
		return fmt.Errorf("peer not found: %s", transfer.PeerID)
	}

	rejectData := FileRejectData{
		TransferID: transferID,
		Reason:     "User rejected",
	}

	data, err := json.Marshal(rejectData)
	if err != nil {
		return fmt.Errorf("marshal reject: %w", err)
	}

	conn, err := transport.Dial(peer.IP, peer.Port)
	if err != nil {
		return fmt.Errorf("connect to peer: %w", err)
	}
	defer conn.Close()

	msg := &transport.Message{
		Type: "file_reject",
		Data: data,
	}

	if err := transport.Send(conn, msg); err != nil {
		return fmt.Errorf("send reject: %w", err)
	}

	log.Printf("[FileTransfer] Rejected file: %s", transfer.FileName)

	if m.events != nil {
		m.events.Emit("file:transfer_rejected", transfer)
	}

	return nil
}

// HandleFileRequest 处理收到的文件请求
func (m *Manager) HandleFileRequest(data json.RawMessage) {
	var req FileRequestData
	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("[FileTransfer] Invalid file request: %v", err)
		return
	}

	transfer := &models.FileTransfer{
		TransferID: req.TransferID,
		PeerID:     req.SenderID,
		Direction:  models.TransferReceiving,
		Status:     models.TransferPending,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		FileHash:   req.FileHash,
		StartedAt:  time.Now(),
	}

	m.mu.Lock()
	m.transfers[req.TransferID] = transfer
	m.mu.Unlock()

	log.Printf("[FileTransfer] Received file request: %s (%d bytes) from %s", req.FileName, req.FileSize, req.SenderID)

	if m.events != nil {
		// 发送方设备信息
		peer := m.peers.Get(req.SenderID)
		eventData := map[string]interface{}{
			"transfer": transfer,
		}
		if peer != nil {
			eventData["sender_name"] = peer.Username + "@" + peer.DeviceName
			eventData["sender_ip"] = peer.IP
			eventData["sender_platform"] = peer.Platform
		}
		m.events.Emit("file:request", eventData)
	}
}

// HandleFileAccept 处理文件接受响应
func (m *Manager) HandleFileAccept(data json.RawMessage) {
	var accept FileAcceptData
	if err := json.Unmarshal(data, &accept); err != nil {
		log.Printf("[FileTransfer] Invalid file accept: %v", err)
		return
	}

	m.mu.Lock()
	transfer, exists := m.transfers[accept.TransferID]
	if !exists {
		m.mu.Unlock()
		log.Printf("[FileTransfer] Transfer not found: %s", accept.TransferID)
		return
	}
	transfer.Status = models.TransferAccepted
	m.mu.Unlock()

	log.Printf("[FileTransfer] File accepted: %s", transfer.FileName)

	// 开始发送文件
	go m.sendFile(transfer, accept.DataPort, accept.ResumeOffset)
}

// HandleFileReject 处理文件拒绝响应
func (m *Manager) HandleFileReject(data json.RawMessage) {
	var reject FileRejectData
	if err := json.Unmarshal(data, &reject); err != nil {
		log.Printf("[FileTransfer] Invalid file reject: %v", err)
		return
	}

	m.mu.Lock()
	transfer, exists := m.transfers[reject.TransferID]
	if !exists {
		m.mu.Unlock()
		return
	}
	transfer.Status = models.TransferRejected
	m.mu.Unlock()

	log.Printf("[FileTransfer] File rejected: %s", transfer.FileName)

	if m.events != nil {
		m.events.Emit("file:transfer_rejected", transfer)
	}
}

// CancelFile 取消文件传输
func (m *Manager) CancelFile(transferID string) error {
	m.mu.Lock()
	transfer, exists := m.transfers[transferID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("transfer not found: %s", transferID)
	}
	transfer.Status = models.TransferCancelled

	// 关闭数据连接以中断传输
	if dataConn, ok := m.dataConns[transferID]; ok {
		dataConn.Close()
		delete(m.dataConns, transferID)
	}
	m.mu.Unlock()

	// 如果是接收方，删除未完成的文件
	if transfer.Direction == models.TransferReceiving && transfer.LocalPath != "" {
		if err := os.Remove(transfer.LocalPath); err != nil && !os.IsNotExist(err) {
			log.Printf("[FileTransfer] Failed to remove incomplete file: %v", err)
		} else {
			log.Printf("[FileTransfer] Removed incomplete file: %s", transfer.LocalPath)
		}
	}

	// 发送取消消息给对端
	peer := m.peers.Get(transfer.PeerID)
	if peer != nil {
		cancelData := FileCancelData{
			TransferID: transferID,
			Reason:     "User cancelled",
		}
		data, err := json.Marshal(cancelData)
		if err == nil {
			conn, err := transport.Dial(peer.IP, peer.Port)
			if err == nil {
				msg := &transport.Message{Type: "file_cancel", Data: data}
				transport.Send(conn, msg)
				conn.Close()
			}
		}
	}

	log.Printf("[FileTransfer] Transfer cancelled: %s", transfer.FileName)

	if m.events != nil {
		m.events.Emit("file:transfer_cancelled", map[string]interface{}{
			"transfer_id": transferID,
			"direction":   transfer.Direction,
		})
	}

	return nil
}

// HandleFileCancel 处理文件取消请求
func (m *Manager) HandleFileCancel(data json.RawMessage) {
	var cancel FileCancelData
	if err := json.Unmarshal(data, &cancel); err != nil {
		log.Printf("[FileTransfer] Invalid file cancel: %v", err)
		return
	}

	m.mu.Lock()
	transfer, exists := m.transfers[cancel.TransferID]
	if !exists {
		m.mu.Unlock()
		return
	}
	transfer.Status = models.TransferCancelled

	// 关闭数据连接
	if dataConn, ok := m.dataConns[cancel.TransferID]; ok {
		dataConn.Close()
		delete(m.dataConns, cancel.TransferID)
	}
	m.mu.Unlock()

	// 如果是接收方，删除未完成的文件
	if transfer.Direction == models.TransferReceiving && transfer.LocalPath != "" {
		if err := os.Remove(transfer.LocalPath); err != nil && !os.IsNotExist(err) {
			log.Printf("[FileTransfer] Failed to remove incomplete file: %v", err)
		} else {
			log.Printf("[FileTransfer] Removed incomplete file: %s", transfer.LocalPath)
		}
	}

	log.Printf("[FileTransfer] Transfer cancelled by peer: %s", transfer.FileName)

	if m.events != nil {
		m.events.Emit("file:transfer_cancelled", map[string]interface{}{
			"transfer_id": cancel.TransferID,
			"direction":   transfer.Direction,
		})
	}
}

// GetTransfer 获取传输状态
func (m *Manager) GetTransfer(transferID string) *models.FileTransfer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.transfers[transferID]
}

// GetAllTransfers 获取所有传输
func (m *Manager) GetAllTransfers() []*models.FileTransfer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transfers := make([]*models.FileTransfer, 0, len(m.transfers))
	for _, t := range m.transfers {
		transfers = append(transfers, t)
	}
	return transfers
}

// sendFile 发送文件
func (m *Manager) sendFile(transfer *models.FileTransfer, dataPort int, resumeOffset int64) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[FileTransfer] PANIC in sendFile: %v", r)
			m.mu.Lock()
			transfer.Status = models.TransferFailed
			m.mu.Unlock()
			if m.events != nil {
				m.events.Emit("file:transfer_failed", transfer)
			}
		}
	}()

	log.Printf("[FileTransfer] Starting to send file: %s", transfer.FileName)

	peer := m.peers.Get(transfer.PeerID)
	if peer == nil {
		log.Printf("[FileTransfer] Peer not found: %s", transfer.PeerID)
		return
	}

	// 连接到接收端的数据端口
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", peer.IP, dataPort), 10*time.Second)
	if err != nil {
		log.Printf("[FileTransfer] Failed to connect to data port: %v", err)
		m.mu.Lock()
		transfer.Status = models.TransferFailed
		m.mu.Unlock()
		if m.events != nil {
			m.events.Emit("file:transfer_failed", transfer)
		}
		return
	}

	m.mu.Lock()
	m.dataConns[transfer.TransferID] = conn
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.dataConns, transfer.TransferID)
		m.mu.Unlock()
		conn.Close()
	}()

	// 创建发送器
	sender := NewSender(&FileTransfer{
		TransferID: transfer.TransferID,
		FileName:   transfer.FileName,
		FileSize:   transfer.FileSize,
		LocalPath:  transfer.LocalPath,
	}, conn, func(transferred, total int64) {
		// 更新进度
		m.mu.Lock()
		transfer.Transferred = transferred
		transfer.Status = models.TransferInProgress
		m.mu.Unlock()

		if m.events != nil {
			m.events.Emit("file:transfer_progress", map[string]interface{}{
				"transfer_id": transfer.TransferID,
				"transferred": transferred,
				"total":       total,
				"progress":    float64(transferred) / float64(total) * 100,
			})
		}
	})

	// 发送文件
	if err := sender.Send(); err != nil {
		log.Printf("[FileTransfer] Failed to send file: %v", err)
		m.mu.Lock()
		transfer.Status = models.TransferFailed
		m.mu.Unlock()
		if m.events != nil {
			m.events.Emit("file:transfer_failed", transfer)
		}
		return
	}

	// 传输完成
	m.mu.Lock()
	transfer.Status = models.TransferCompleted
	transfer.CompletedAt = time.Now()
	m.mu.Unlock()

	if m.events != nil {
		m.events.Emit("file:transfer_completed", transfer)
	}

	if m.chat != nil {
		m.chat.AddFileMessage(transfer.FileName, transfer.LocalPath, transfer.FileSize, m.deviceID, m.deviceName, true)
	}

	log.Printf("[FileTransfer] File sent successfully: %s", transfer.FileName)
}

// receiveFile 接收文件
func (m *Manager) receiveFile(transfer *models.FileTransfer) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[FileTransfer] PANIC in receiveFile: %v", r)
			m.mu.Lock()
			transfer.Status = models.TransferFailed
			m.mu.Unlock()
			if m.events != nil {
				m.events.Emit("file:transfer_failed", transfer)
			}
		}
	}()

	log.Printf("[FileTransfer] Starting to receive file: %s", transfer.FileName)

	// 监听数据端口
	listener, err := net.Listen("tcp", ":0") // 随机端口
	if err != nil {
		log.Printf("[FileTransfer] Failed to listen: %v", err)
		m.mu.Lock()
		transfer.Status = models.TransferFailed
		m.mu.Unlock()
		if m.events != nil {
			m.events.Emit("file:transfer_failed", transfer)
		}
		return
	}
	defer listener.Close()

	dataPort := listener.Addr().(*net.TCPAddr).Port
	log.Printf("[FileTransfer] Listening on port %d for file data", dataPort)

	// 发送接受响应，包含数据端口
	peer := m.peers.Get(transfer.PeerID)
	if peer == nil {
		log.Printf("[FileTransfer] Peer not found: %s", transfer.PeerID)
		return
	}

	acceptData := FileAcceptData{
		TransferID:   transfer.TransferID,
		DataPort:     dataPort,
		ResumeOffset: 0,
	}

	data, err := json.Marshal(acceptData)
	if err != nil {
		log.Printf("[FileTransfer] Failed to marshal accept: %v", err)
		return
	}

	conn, err := transport.Dial(peer.IP, peer.Port)
	if err != nil {
		log.Printf("[FileTransfer] Failed to connect to peer: %v", err)
		return
	}
	defer conn.Close()

	msg := &transport.Message{
		Type: "file_accept",
		Data: data,
	}

	if err := transport.Send(conn, msg); err != nil {
		log.Printf("[FileTransfer] Failed to send accept: %v", err)
		return
	}

	// 等待发送端连接
	dataConn, err := listener.Accept()
	if err != nil {
		log.Printf("[FileTransfer] Failed to accept connection: %v", err)
		m.mu.Lock()
		transfer.Status = models.TransferFailed
		m.mu.Unlock()
		if m.events != nil {
			m.events.Emit("file:transfer_failed", transfer)
		}
		return
	}

	m.mu.Lock()
	m.dataConns[transfer.TransferID] = dataConn
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.dataConns, transfer.TransferID)
		m.mu.Unlock()
		dataConn.Close()
	}()

	// 创建接收器
	receiver := NewReceiver(&FileTransfer{
		TransferID: transfer.TransferID,
		FileName:   transfer.FileName,
		FileSize:   transfer.FileSize,
	}, dataConn, transfer.LocalPath, func(transferred, total int64) {
		// 更新进度
		m.mu.Lock()
		transfer.Transferred = transferred
		transfer.Status = models.TransferInProgress
		m.mu.Unlock()

		if m.events != nil {
			m.events.Emit("file:transfer_progress", map[string]interface{}{
				"transfer_id": transfer.TransferID,
				"transferred": transferred,
				"total":       total,
				"progress":    float64(transferred) / float64(total) * 100,
			})
		}
	})

	// 接收文件
	if err := receiver.Receive(); err != nil {
		log.Printf("[FileTransfer] Failed to receive file: %v", err)
		m.mu.Lock()
		transfer.Status = models.TransferFailed
		m.mu.Unlock()
		if m.events != nil {
			m.events.Emit("file:transfer_failed", transfer)
		}
		return
	}

	// 传输完成
	m.mu.Lock()
	transfer.Status = models.TransferCompleted
	transfer.CompletedAt = time.Now()
	m.mu.Unlock()

	if m.events != nil {
		m.events.Emit("file:transfer_completed", transfer)
	}

	if m.chat != nil {
		peerName := transfer.PeerID
		if peer := m.peers.Get(transfer.PeerID); peer != nil {
			peerName = peer.DeviceName
		}
		m.chat.AddFileMessage(transfer.FileName, transfer.LocalPath, transfer.FileSize, transfer.PeerID, peerName, false)
	}

	log.Printf("[FileTransfer] File received successfully: %s", transfer.FileName)
}

// 消息数据结构
type FileRequestData struct {
	TransferID string `json:"transfer_id"`
	SenderID   string `json:"sender_id"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	FileHash   string `json:"file_hash"`
	MimeType   string `json:"mime_type"`
}

type FileAcceptData struct {
	TransferID   string `json:"transfer_id"`
	DataPort     int    `json:"data_port"`
	ResumeOffset int64  `json:"resume_offset"`
}

type FileRejectData struct {
	TransferID string `json:"transfer_id"`
	Reason     string `json:"reason"`
}

type FileCancelData struct {
	TransferID string `json:"transfer_id"`
	Reason     string `json:"reason"`
}

type FileInfo struct {
	Name     string
	Size     int64
	Hash     string
	MimeType string
}

func getFileInfo(filePath string) (*FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	// 计算 MD5
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("calculate md5: %w", err)
	}
	hashSum := hex.EncodeToString(hash.Sum(nil))

	// 获取文件名
	fileName := filepath.Base(filePath)

	// 简单的 MIME 类型检测
	mimeType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".txt", ".md", ".log":
		mimeType = "text/plain"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".pdf":
		mimeType = "application/pdf"
	case ".zip":
		mimeType = "application/zip"
	case ".mp3":
		mimeType = "audio/mpeg"
	case ".mp4":
		mimeType = "video/mp4"
	}

	return &FileInfo{
		Name:     fileName,
		Size:     stat.Size(),
		Hash:     hashSum,
		MimeType: mimeType,
	}, nil
}
