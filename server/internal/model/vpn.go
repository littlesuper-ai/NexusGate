package model

import (
	"time"

	"gorm.io/gorm"
)

type WireGuardInterface struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	DeviceID   uint           `json:"device_id" gorm:"index;not null"`
	Name       string         `json:"name" gorm:"not null"`       // e.g. wg0
	PrivateKey string         `json:"-" gorm:"not null"`
	PublicKey  string         `json:"public_key"`
	Address    string         `json:"address"`                     // e.g. 10.99.0.1/24
	ListenPort int            `json:"listen_port" gorm:"default:51820"`
	Enabled    bool           `json:"enabled" gorm:"default:true"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

type WireGuardPeer struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	InterfaceID   uint           `json:"interface_id" gorm:"index;not null"`
	Description   string         `json:"description"`
	PublicKey     string         `json:"public_key" gorm:"not null"`
	PresharedKey  string         `json:"-"`
	AllowedIPs    string         `json:"allowed_ips"`    // comma-separated CIDRs
	Endpoint      string         `json:"endpoint"`       // host:port
	Keepalive     int            `json:"keepalive" gorm:"default:25"`
	Enabled       bool           `json:"enabled" gorm:"default:true"`
	LastHandshake *time.Time     `json:"last_handshake"`
	TxBytes       int64          `json:"tx_bytes"`
	RxBytes       int64          `json:"rx_bytes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
