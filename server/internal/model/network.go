package model

import (
	"time"

	"gorm.io/gorm"
)

// --- Multi-WAN (mwan3) ---

type WANInterface struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	DeviceID    uint           `json:"device_id" gorm:"index;not null"`
	Name        string         `json:"name" gorm:"not null"`        // wan1, wan2
	Interface   string         `json:"interface" gorm:"not null"`    // eth1, pppoe-wan
	Enabled     bool           `json:"enabled" gorm:"default:true"`
	Weight      int            `json:"weight" gorm:"default:1"`
	TrackIPs    string         `json:"track_ips"`                    // comma-separated: 8.8.8.8,114.114.114.114
	Reliability int            `json:"reliability" gorm:"default:2"`
	Interval    int            `json:"interval" gorm:"default:5"`    // probe interval secs
	Down        int            `json:"down" gorm:"default:3"`        // failures before mark down
	Up          int            `json:"up" gorm:"default:3"`          // successes before mark up
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type MWANPolicy struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	DeviceID  uint           `json:"device_id" gorm:"index;not null"`
	Name      string         `json:"name" gorm:"not null"`
	Members   string         `json:"members"`     // JSON: [{"iface":"wan1","metric":1,"weight":1}]
	LastResort string        `json:"last_resort" gorm:"default:default"` // default, unreachable
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type MWANRule struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	DeviceID  uint           `json:"device_id" gorm:"index;not null"`
	Name      string         `json:"name" gorm:"not null"`
	SrcIP     string         `json:"src_ip"`
	DestIP    string         `json:"dest_ip"`
	Proto     string         `json:"proto"`       // tcp, udp, all
	SrcPort   string         `json:"src_port"`
	DestPort  string         `json:"dest_port"`
	Policy    string         `json:"policy" gorm:"not null"` // policy name to use
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	Position  int            `json:"position" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// --- DHCP ---

type DHCPPool struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	DeviceID  uint           `json:"device_id" gorm:"index;not null"`
	Interface string         `json:"interface" gorm:"not null"` // lan, guest, iot
	Start     int            `json:"start" gorm:"default:100"`
	Limit     int            `json:"limit" gorm:"default:150"`
	LeaseTime string         `json:"lease_time" gorm:"default:12h"`
	DNS       string         `json:"dns"`         // comma-separated DNS servers
	Gateway   string         `json:"gateway"`
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type StaticLease struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	DeviceID  uint           `json:"device_id" gorm:"index;not null"`
	Name      string         `json:"name" gorm:"not null"`
	MAC       string         `json:"mac" gorm:"not null"`
	IP        string         `json:"ip" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// --- VLAN ---

type VLAN struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	DeviceID  uint           `json:"device_id" gorm:"index;not null"`
	VID       int            `json:"vid" gorm:"not null"`           // 802.1Q VLAN ID
	Name      string         `json:"name" gorm:"not null"`          // office, server, guest
	Interface string         `json:"interface"`                      // br-lan.10
	IPAddr    string         `json:"ip_addr"`                        // 10.0.10.1
	Netmask   string         `json:"netmask" gorm:"default:255.255.255.0"`
	Isolated  bool           `json:"isolated" gorm:"default:false"` // inter-VLAN routing disabled
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
