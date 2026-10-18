package models

import "time"

type MessageType string

const (
	MsgTypeChat   MessageType = "chat"
	MsgTypeSystem MessageType = "system"
	MsgTypeFile   MessageType = "file"
)

type ChatMessage struct {
	ID         string      `json:"id"`
	Type       MessageType `json:"type"`
	Content    string      `json:"content"`
	SenderID   string      `json:"sender_id"`
	SenderName string      `json:"sender_name"`
	Timestamp  time.Time   `json:"timestamp"`
	FilePath   string      `json:"file_path,omitempty"`
	FileName   string      `json:"file_name,omitempty"`
	FileSize   int64       `json:"file_size,omitempty"`
}
