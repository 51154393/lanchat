package transport

import (
	"context"
	"fmt"
	"log"
	"net"
)

type MessageHandler func(conn net.Conn, msg *Message)

type Server struct {
	listener net.Listener
	port     int
	handler  MessageHandler
}

func NewServer(port int, handler MessageHandler) *Server {
	return &Server{
		port:    port,
		handler: handler,
	}
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.listener = listener
	log.Printf("[TCP] Server listening on port %d", s.port)

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("[TCP] Accept error: %v", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		msg, err := Decode(conn)
		if err != nil {
			log.Printf("[TCP] Decode error from %s: %v", conn.RemoteAddr(), err)
			return
		}

		if s.handler != nil {
			s.handler(conn, msg)
		}
	}
}

func (s *Server) Port() int {
	if s.listener != nil {
		return s.listener.Addr().(*net.TCPAddr).Port
	}
	return s.port
}
