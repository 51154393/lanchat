package chat

import (
	"context"
	"encoding/json"
	"lanchat/internal/models"
	"lanchat/internal/transport"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ChatMessageData struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
	Sender     string `json:"sender"`
	SenderName string `json:"sender_name"`
}

type HeartbeatData struct {
	Timestamp int64 `json:"timestamp"`
}

type EventEmitter interface {
	Emit(eventName string, data interface{})
}

type PeerConnector interface {
	GetOnline() []*models.Device
	UpdateLastSeen(deviceID string)
}

type Service struct {
	deviceID   string
	deviceName string
	peers      PeerConnector
	events     EventEmitter
	messages   []*models.ChatMessage
	mu         sync.RWMutex
}

func NewService(deviceID, deviceName string, peers PeerConnector, events EventEmitter) *Service {
	return &Service{
		deviceID:   deviceID,
		deviceName: deviceName,
		peers:      peers,
		events:     events,
		messages:   make([]*models.ChatMessage, 0),
	}
}

func (s *Service) Send(ctx context.Context, content string) error {
	msg := ChatMessageData{
		ID:         uuid.New().String(),
		Content:    content,
		Timestamp:  time.Now().UnixMilli(),
		Sender:     s.deviceID,
		SenderName: s.deviceName,
	}

	chatMsg := &models.ChatMessage{
		ID:         msg.ID,
		Type:       models.MsgTypeChat,
		Content:    content,
		SenderID:   msg.Sender,
		SenderName: msg.SenderName,
		Timestamp:  time.UnixMilli(msg.Timestamp),
	}

	s.storeMessage(chatMsg)

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	transportMsg := &transport.Message{
		Type: "chat",
		Data: data,
	}

	onlinePeers := s.peers.GetOnline()
	for _, peer := range onlinePeers {
		if peer.DeviceID == s.deviceID {
			continue
		}
		go func(p *models.Device) {
			conn, err := transport.Dial(p.IP, p.Port)
			if err != nil {
				log.Printf("[Chat] Failed to connect to %s: %v", p.DeviceName, err)
				return
			}
			defer conn.Close()

			if err := transport.Send(conn, transportMsg); err != nil {
				log.Printf("[Chat] Failed to send to %s: %v", p.DeviceName, err)
			}
		}(peer)
	}

	return nil
}

func (s *Service) HandleMessage(conn net.Conn, msg *transport.Message) {
	switch msg.Type {
	case "chat":
		var data ChatMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Printf("[Chat] Invalid chat message: %v", err)
			return
		}

		chatMsg := &models.ChatMessage{
			ID:         data.ID,
			Type:       models.MsgTypeChat,
			Content:    data.Content,
			SenderID:   data.Sender,
			SenderName: data.SenderName,
			Timestamp:  time.UnixMilli(data.Timestamp),
		}

		s.storeMessage(chatMsg)
		if s.events != nil {
			s.events.Emit("chat:message", chatMsg)
		}

	case "heartbeat":
		var data HeartbeatData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Printf("[Chat] Invalid heartbeat: %v", err)
			return
		}
		// 从连接中获取对端信息并更新最后活跃时间
		// 这里简化处理，实际需要从连接中解析出 device_id

	case "file_request":
		if s.events != nil {
			s.events.Emit("file:request", msg.Data)
		}

	case "file_accept":
		if s.events != nil {
			s.events.Emit("file:accept", msg.Data)
		}

	case "file_reject":
		if s.events != nil {
			s.events.Emit("file:reject", msg.Data)
		}
	}
}

func (s *Service) storeMessage(msg *models.ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

func (s *Service) GetMessages() []*models.ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.ChatMessage, len(s.messages))
	copy(result, s.messages)
	return result
}

// AddFileMessage 添加文件传输完成的消息到聊天记录
func (s *Service) AddFileMessage(fileName, filePath string, fileSize int64, senderID, senderName string, isSender bool) {
	content := "发送了文件"
	if !isSender {
		content = "接收了文件"
	}

	msg := &models.ChatMessage{
		ID:         uuid.New().String(),
		Type:       models.MsgTypeFile,
		Content:    content,
		SenderID:   senderID,
		SenderName: senderName,
		Timestamp:  time.Now(),
		FilePath:   filePath,
		FileName:   fileName,
		FileSize:   fileSize,
	}

	s.storeMessage(msg)
	if s.events != nil {
		s.events.Emit("chat:message", msg)
	}
}

func (s *Service) SendHeartbeat(ctx context.Context) error {
	data, err := json.Marshal(HeartbeatData{
		Timestamp: time.Now().UnixMilli(),
	})
	if err != nil {
		return err
	}

	msg := &transport.Message{
		Type: "heartbeat",
		Data: data,
	}

	onlinePeers := s.peers.GetOnline()
	for _, peer := range onlinePeers {
		if peer.DeviceID == s.deviceID {
			continue
		}
		go func(p *models.Device) {
			conn, err := transport.Dial(p.IP, p.Port)
			if err != nil {
				return
			}
			defer conn.Close()
			transport.Send(conn, msg)
		}(peer)
	}

	return nil
}
