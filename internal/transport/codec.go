package transport

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

const (
	MaxPayloadSize = 1 << 20 // 1 MB
	HeaderSize     = 8
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func Encode(msg *Message) ([]byte, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}

	if len(payload) > MaxPayloadSize {
		return nil, fmt.Errorf("payload too large: %d > %d", len(payload), MaxPayloadSize)
	}

	buf := make([]byte, HeaderSize+len(payload))
	buf[0] = 0x01 // Type: JSON
	buf[1] = 0x00 // Reserved
	buf[2] = 0x00 // Reserved
	buf[3] = 0x00 // Reserved
	binary.BigEndian.PutUint32(buf[4:8], uint32(len(payload)))
	copy(buf[8:], payload)

	return buf, nil
}

func Decode(r io.Reader) (*Message, error) {
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	payloadLen := binary.BigEndian.Uint32(header[4:8])
	if payloadLen > MaxPayloadSize {
		return nil, fmt.Errorf("payload too large: %d > %d", payloadLen, MaxPayloadSize)
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal message: %w", err)
	}

	return &msg, nil
}
