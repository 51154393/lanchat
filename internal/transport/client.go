package transport

import (
	"fmt"
	"net"
	"time"
)

func Dial(address string, port int) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", address, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return conn, nil
}

func Send(conn net.Conn, msg *Message) error {
	data, err := Encode(msg)
	if err != nil {
		return fmt.Errorf("encode message: %w", err)
	}

	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}
