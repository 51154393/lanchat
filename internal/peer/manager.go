package peer

import (
	"context"
	"lanchat/internal/models"
	"log"
	"sync"
	"time"
)

const (
	heartbeatInterval = 10 * time.Second
	heartbeatTimeout  = 30 * time.Second
	deviceExpireTime  = 60 * time.Second
)

type EventEmitter interface {
	Emit(eventName string, data interface{})
}

type Manager struct {
	peers  map[string]*models.Device
	mu     sync.RWMutex
	events EventEmitter
}

func NewManager(events EventEmitter) *Manager {
	return &Manager{
		peers:  make(map[string]*models.Device),
		events: events,
	}
}

func (m *Manager) AddOrUpdate(device *models.Device) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.peers[device.DeviceID]
	if exists {
		existing.IP = device.IP
		existing.Port = device.Port
		existing.Status = models.StatusOnline
		existing.LastSeen = time.Now()
		log.Printf("[Peer] Updated: %s@%s (%s)", existing.Username, existing.DeviceName, existing.IP)
	} else {
		device.Status = models.StatusOnline
		device.LastSeen = time.Now()
		m.peers[device.DeviceID] = device
		log.Printf("[Peer] Added: %s@%s (%s)", device.Username, device.DeviceName, device.IP)
		if m.events != nil {
			m.events.Emit("device:joined", device)
		}
	}
}

func (m *Manager) Remove(deviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if peer, exists := m.peers[deviceID]; exists {
		log.Printf("[Peer] Removed: %s@%s", peer.Username, peer.DeviceName)
		delete(m.peers, deviceID)
		if m.events != nil {
			m.events.Emit("device:left", peer)
		}
	}
}

func (m *Manager) Get(deviceID string) *models.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.peers[deviceID]
}

func (m *Manager) GetAll() []*models.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*models.Device, 0, len(m.peers))
	for _, d := range m.peers {
		devices = append(devices, d)
	}
	return devices
}

func (m *Manager) GetOnline() []*models.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*models.Device, 0)
	for _, d := range m.peers {
		if d.Status == models.StatusOnline {
			devices = append(devices, d)
		}
	}
	return devices
}

func (m *Manager) GetOnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, d := range m.peers {
		if d.Status == models.StatusOnline {
			count++
		}
	}
	return count
}

func (m *Manager) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkTimeouts()
		}
	}
}

func (m *Manager) checkTimeouts() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, peer := range m.peers {
		if peer.Status == models.StatusOnline && now.Sub(peer.LastSeen) > heartbeatTimeout {
			peer.Status = models.StatusOffline
			log.Printf("[Peer] Offline: %s@%s", peer.Username, peer.DeviceName)
			if m.events != nil {
				m.events.Emit("device:offline", peer)
			}
		}
		if peer.Status == models.StatusOffline && now.Sub(peer.LastSeen) > deviceExpireTime {
			log.Printf("[Peer] Expired: %s@%s", peer.Username, peer.DeviceName)
			delete(m.peers, id)
			if m.events != nil {
				m.events.Emit("device:left", peer)
			}
		}
	}
}

func (m *Manager) UpdateLastSeen(deviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if peer, exists := m.peers[deviceID]; exists {
		peer.LastSeen = time.Now()
		if peer.Status == models.StatusOffline {
			peer.Status = models.StatusOnline
			if m.events != nil {
				m.events.Emit("device:online", peer)
			}
		}
	}
}
