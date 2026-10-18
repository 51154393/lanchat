package filetransfer

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// Sender 文件发送器
type Sender struct {
	transfer    *FileTransfer
	conn        net.Conn
	chunkSize   int
	onProgress  func(transferred int64, total int64)
}

func NewSender(transfer *FileTransfer, conn net.Conn, onProgress func(int64, int64)) *Sender {
	return &Sender{
		transfer:   transfer,
		conn:       conn,
		chunkSize:  65536, // 64KB
		onProgress: onProgress,
	}
}

func (s *Sender) Send() error {
	file, err := os.Open(s.transfer.LocalPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// 发送文件头
	if err := s.sendHeader(file); err != nil {
		return fmt.Errorf("send header: %w", err)
	}

	// 发送数据块
	reader := bufio.NewReader(file)
	buffer := make([]byte, s.chunkSize)
	chunkIndex := uint32(0)
	transferred := int64(0)

	for {
		n, err := reader.Read(buffer)
		if n == 0 {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read chunk: %w", err)
		}

		chunk := buffer[:n]
		chunkMD5 := md5.Sum(chunk)

		// 发送数据块：ChunkIndex + ChunkLen + MD5 + Data
		if err := binary.Write(s.conn, binary.BigEndian, chunkIndex); err != nil {
			return fmt.Errorf("send chunk index: %w", err)
		}
		if err := binary.Write(s.conn, binary.BigEndian, uint32(n)); err != nil {
			return fmt.Errorf("send chunk len: %w", err)
		}
		if _, err := s.conn.Write(chunkMD5[:]); err != nil {
			return fmt.Errorf("send chunk md5: %w", err)
		}
		if _, err := s.conn.Write(chunk); err != nil {
			return fmt.Errorf("send chunk data: %w", err)
		}

		// 等待确认
		var ackIndex uint32
		var ackStatus byte
		if err := binary.Read(s.conn, binary.BigEndian, &ackIndex); err != nil {
			return fmt.Errorf("read ack index: %w", err)
		}
		if err := binary.Read(s.conn, binary.BigEndian, &ackStatus); err != nil {
			return fmt.Errorf("read ack status: %w", err)
		}

		if ackStatus != 0 {
			return fmt.Errorf("chunk %d rejected", chunkIndex)
		}

		transferred += int64(n)
		chunkIndex++

		if s.onProgress != nil {
			s.onProgress(transferred, s.transfer.FileSize)
		}
	}

	// 发送完成标记
	if _, err := s.conn.Write([]byte("DONE")); err != nil {
		return fmt.Errorf("send done: %w", err)
	}

	log.Printf("[FileTransfer] File sent successfully: %s", s.transfer.FileName)
	return nil
}

func (s *Sender) sendHeader(file *os.File) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// 文件头：Magic(4) + Version(1) + TransferID(36) + FileNameLen(2) + FileName + TotalSize(8) + ChunkSize(4) + Offset(8)
	header := []byte("LNFT")                          // Magic
	header = append(header, 0x01)                      // Version
	header = append(header, []byte(s.transfer.TransferID)...) // TransferID (36 bytes)
	
	fileNameBytes := []byte(s.transfer.FileName)
	nameLen := uint16(len(fileNameBytes))
	nameLenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(nameLenBytes, nameLen)
	header = append(header, nameLenBytes...)
	header = append(header, fileNameBytes...)

	totalSizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(totalSizeBytes, uint64(stat.Size()))
	header = append(header, totalSizeBytes...)

	chunkSizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(chunkSizeBytes, uint32(s.chunkSize))
	header = append(header, chunkSizeBytes...)

	offsetBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(offsetBytes, 0) // 暂不支持断点续传
	header = append(header, offsetBytes...)

	_, err = s.conn.Write(header)
	return err
}

// Receiver 文件接收器
type Receiver struct {
	transfer   *FileTransfer
	conn       net.Conn
	savePath   string
	onProgress func(transferred int64, total int64)
}

func NewReceiver(transfer *FileTransfer, conn net.Conn, savePath string, onProgress func(int64, int64)) *Receiver {
	return &Receiver{
		transfer:   transfer,
		conn:       conn,
		savePath:   savePath,
		onProgress: onProgress,
	}
}

func (r *Receiver) Receive() error {
	// 读取文件头
	header := make([]byte, 4+1+36+2) // Magic + Version + TransferID + FileNameLen
	if _, err := io.ReadFull(r.conn, header); err != nil {
		return fmt.Errorf("read header start: %w", err)
	}

	// 验证 Magic
	if string(header[:4]) != "LNFT" {
		return fmt.Errorf("invalid magic")
	}

	// 读取文件名长度
	nameLen := binary.BigEndian.Uint16(header[41:43])

	// 读取文件名
	nameBytes := make([]byte, nameLen)
	if _, err := io.ReadFull(r.conn, nameBytes); err != nil {
		return fmt.Errorf("read filename: %w", err)
	}
	fileName := string(nameBytes)

	// 读取 TotalSize + ChunkSize + Offset
	meta := make([]byte, 8+4+8)
	if _, err := io.ReadFull(r.conn, meta); err != nil {
		return fmt.Errorf("read meta: %w", err)
	}
	totalSize := int64(binary.BigEndian.Uint64(meta[0:8]))
	chunkSize := int(binary.BigEndian.Uint32(meta[8:12]))
	// offset := int64(binary.BigEndian.Uint64(meta[12:20]))

	log.Printf("[FileTransfer] Receiving file: %s (%d bytes)", fileName, totalSize)

	// 创建文件
	file, err := os.Create(r.savePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	transferred := int64(0)

	// 计算预期的数据块数量
	totalChunks := uint32(0)
	if totalSize > 0 {
		totalChunks = uint32((totalSize + int64(chunkSize) - 1) / int64(chunkSize))
	}

	// 接收数据块
	for i := uint32(0); i < totalChunks; i++ {
		var chunkIndex uint32
		var chunkLen uint32
		chunkMD5 := make([]byte, 16)

		// 读取 ChunkIndex
		if err := binary.Read(r.conn, binary.BigEndian, &chunkIndex); err != nil {
			return fmt.Errorf("read chunk index: %w", err)
		}

		// 读取 ChunkLen
		if err := binary.Read(r.conn, binary.BigEndian, &chunkLen); err != nil {
			return fmt.Errorf("read chunk len: %w", err)
		}

		// 读取 MD5
		if _, err := io.ReadFull(r.conn, chunkMD5); err != nil {
			return fmt.Errorf("read chunk md5: %w", err)
		}

		// 读取数据
		chunk := make([]byte, chunkLen)
		if _, err := io.ReadFull(r.conn, chunk); err != nil {
			return fmt.Errorf("read chunk data: %w", err)
		}

		// 验证 MD5
		actualMD5 := md5.Sum(chunk)
		if !bytesEqual(chunkMD5, actualMD5[:]) {
			binary.Write(r.conn, binary.BigEndian, chunkIndex)
			binary.Write(r.conn, binary.BigEndian, byte(1))
			return fmt.Errorf("chunk %d md5 mismatch", chunkIndex)
		}

		// 写入文件
		if _, err := writer.Write(chunk); err != nil {
			return fmt.Errorf("write chunk: %w", err)
		}

		// 发送确认
		if err := binary.Write(r.conn, binary.BigEndian, chunkIndex); err != nil {
			return fmt.Errorf("send ack index: %w", err)
		}
		if err := binary.Write(r.conn, binary.BigEndian, byte(0)); err != nil {
			return fmt.Errorf("send ack status: %w", err)
		}

		transferred += int64(chunkLen)
		if r.onProgress != nil {
			r.onProgress(transferred, totalSize)
		}
	}

	// 所有数据块接收完毕，读取 DONE 标记
	doneMarker := make([]byte, 4)
	if _, err := io.ReadFull(r.conn, doneMarker); err != nil {
		return fmt.Errorf("read done marker: %w", err)
	}
	if string(doneMarker) != "DONE" {
		return fmt.Errorf("invalid done marker")
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush file: %w", err)
	}

	log.Printf("[FileTransfer] File received successfully: %s", fileName)
	return nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// FileTransfer 传输状态（从 models 包复制，避免循环依赖）
type FileTransfer struct {
	TransferID  string
	PeerID      string
	Direction   int
	Status      int
	FileName    string
	FileSize    int64
	FileHash    string
	LocalPath   string
	ChunkSize   int
}
