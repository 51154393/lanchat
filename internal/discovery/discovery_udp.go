package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"lanchat/internal/config"
	"lanchat/internal/models"
	"log"
	"net"
	"time"
)

const (
	broadcastPort     = 9528
	broadcastInterval = 5 * time.Second
	bufferSize        = 1024
)

type BroadcastMessage struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Username   string `json:"username"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Platform   string `json:"platform"`
}

type Discovery struct {
	cfg           *config.Config
	broadcastConn *net.UDPConn
	listenConn    *net.UDPConn
}

func New(cfg *config.Config) *Discovery {
	return &Discovery{cfg: cfg}
}

func (d *Discovery) Register(ctx context.Context) error {
	log.Printf("[UDP] Starting broadcast discovery on port %d", broadcastPort)

	// 设置广播连接
	broadcastAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", broadcastPort))
	if err != nil {
		return fmt.Errorf("resolve broadcast address: %w", err)
	}

	broadcastConn, err := net.DialUDP("udp4", nil, broadcastAddr)
	if err != nil {
		return fmt.Errorf("dial broadcast: %w", err)
	}
	d.broadcastConn = broadcastConn

	// 设置监听连接
	listenAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", broadcastPort))
	if err != nil {
		return fmt.Errorf("resolve listen address: %w", err)
	}

	listenConn, err := net.ListenUDP("udp4", listenAddr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	d.listenConn = listenConn

	log.Printf("[UDP] ✓ Broadcast discovery initialized")

	// 启动广播发送
	go d.startBroadcast(ctx)

	return nil
}

func (d *Discovery) Browse(ctx context.Context, onFound func(*models.Device)) error {
	log.Printf("[UDP] Starting broadcast listener")

	// 启动广播接收
	go d.startListening(ctx, onFound)

	return nil
}

func (d *Discovery) startBroadcast(ctx context.Context) {
	ticker := time.NewTicker(broadcastInterval)
	defer ticker.Stop()

	// 立即发送一次
	d.sendBroadcast()
	log.Printf("[UDP] First broadcast sent")

	for {
		select {
		case <-ctx.Done():
			log.Printf("[UDP] Stopping broadcast")
			return
		case <-ticker.C:
			d.sendBroadcast()
		}
	}
}

func (d *Discovery) sendBroadcast() {
	msg := BroadcastMessage{
		DeviceID:   d.cfg.DeviceID,
		DeviceName: d.cfg.DeviceName,
		Username:   d.cfg.Username,
		IP:         d.getLocalIP(),
		Port:       d.cfg.TCPListenPort,
		Platform:   config.GetPlatform(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[UDP] Marshal error: %v", err)
		return
	}

	_, err = d.broadcastConn.Write(data)
	if err != nil {
		log.Printf("[UDP] Broadcast send error: %v", err)
	}
}

func (d *Discovery) startListening(ctx context.Context, onFound func(*models.Device)) {
	buffer := make([]byte, bufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			d.listenConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := d.listenConn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("[UDP] Read error from %s: %v", addr, err)
				continue
			}

			var msg BroadcastMessage
			if err := json.Unmarshal(buffer[:n], &msg); err != nil {
				log.Printf("[UDP] Unmarshal error: %v", err)
				continue
			}

			// 忽略自己的广播
			if msg.DeviceID == d.cfg.DeviceID {
				continue
			}

			log.Printf("[UDP] ✓ Discovered peer: %s@%s at %s:%d",
				msg.Username, msg.DeviceName, msg.IP, msg.Port)

			if onFound != nil {
				onFound(&models.Device{
					DeviceID:   msg.DeviceID,
					DeviceName: msg.DeviceName,
					Username:   msg.Username,
					IP:         msg.IP,
					Port:       msg.Port,
					Platform:   msg.Platform,
					Status:     models.StatusOnline,
				})
			}
		}
	}
}

func (d *Discovery) getLocalIP() string {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		log.Printf("[UDP] Failed to get local IP: %v", err)
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (d *Discovery) Shutdown() {
	if d.broadcastConn != nil {
		d.broadcastConn.Close()
	}
	if d.listenConn != nil {
		d.listenConn.Close()
	}
}
