package models

import "time"

type TransferStatus int

const (
	TransferPending TransferStatus = iota
	TransferAccepted
	TransferRejected
	TransferInProgress
	TransferCompleted
	TransferFailed
	TransferCancelled
)

type TransferDirection int

const (
	TransferSending TransferDirection = iota
	TransferReceiving
)

type FileTransfer struct {
	TransferID   string            `json:"transfer_id"`
	PeerID       string            `json:"peer_id"`
	Direction    TransferDirection `json:"direction"`
	Status       TransferStatus    `json:"status"`
	FileName     string            `json:"file_name"`
	FileSize     int64             `json:"file_size"`
	FileHash     string            `json:"file_hash"`
	LocalPath    string            `json:"local_path"`
	Transferred  int64             `json:"transferred"`
	ResumeOffset int64             `json:"resume_offset"`
	ChunkSize    int               `json:"chunk_size"`
	StartedAt    time.Time         `json:"started_at"`
	CompletedAt  time.Time         `json:"completed_at"`
}

func (t *FileTransfer) Progress() float64 {
	if t.FileSize == 0 {
		return 0
	}
	return float64(t.Transferred) / float64(t.FileSize) * 100
}
