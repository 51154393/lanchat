package models

import "time"

type DeviceStatus int

const (
	StatusOffline DeviceStatus = iota
	StatusOnline
)

type Device struct {
	DeviceID   string       `json:"device_id"`
	DeviceName string       `json:"device_name"`
	Username   string       `json:"username"`
	IP         string       `json:"ip"`
	Port       int          `json:"port"`
	Platform   string       `json:"platform"`
	Status     DeviceStatus `json:"status"`
	LastSeen   time.Time    `json:"last_seen"`
}
