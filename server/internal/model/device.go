package model

import (
	"time"

	"gorm.io/gorm"
)

type DeviceStatus string

const (
	StatusOnline  DeviceStatus = "online"
	StatusOffline DeviceStatus = "offline"
	StatusUnknown DeviceStatus = "unknown"
)

type Device struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"not null"`
	MAC          string         `json:"mac" gorm:"uniqueIndex;not null"`
	IPAddress    string         `json:"ip_address"`
	Model        string         `json:"model"`
	Firmware     string         `json:"firmware"`
	Status       DeviceStatus   `json:"status" gorm:"default:unknown"`
	Group        string         `json:"group" gorm:"index"`
	Tags         string         `json:"tags"`
	UptimeSecs   int64          `json:"uptime_secs"`
	CPUUsage     float64        `json:"cpu_usage"`
	MemUsage     float64        `json:"mem_usage"`
	LastSeenAt   *time.Time     `json:"last_seen_at"`
	RegisteredAt time.Time      `json:"registered_at" gorm:"autoCreateTime"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type DeviceMetrics struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DeviceID   uint      `json:"device_id" gorm:"index;not null"`
	CPUUsage   float64   `json:"cpu_usage"`
	MemUsage   float64   `json:"mem_usage"`
	MemTotal   int64     `json:"mem_total"`
	MemFree    int64     `json:"mem_free"`
	RxBytes    int64     `json:"rx_bytes"`
	TxBytes    int64     `json:"tx_bytes"`
	Conntrack  int       `json:"conntrack"`
	UptimeSecs int64     `json:"uptime_secs"`
	LoadAvg    string    `json:"load_avg"`
	CollectedAt time.Time `json:"collected_at"`
}
