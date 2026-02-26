package model

import (
	"time"

	"gorm.io/gorm"
)

type FirewallZone struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	DeviceID  uint           `json:"device_id" gorm:"uniqueIndex:idx_zone_device_name;not null"`
	Name      string         `json:"name" gorm:"uniqueIndex:idx_zone_device_name;not null"`
	Input     string         `json:"input" gorm:"default:REJECT"`  // ACCEPT, REJECT, DROP
	Output    string         `json:"output" gorm:"default:ACCEPT"`
	Forward   string         `json:"forward" gorm:"default:REJECT"`
	Masq      bool           `json:"masq"`
	Networks  string         `json:"networks"` // comma-separated: lan,guest
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type FirewallRule struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	DeviceID  uint           `json:"device_id" gorm:"uniqueIndex:idx_rule_device_name;not null"`
	Name      string         `json:"name" gorm:"uniqueIndex:idx_rule_device_name;not null"`
	Src       string         `json:"src"`       // source zone
	Dest      string         `json:"dest"`      // dest zone
	Proto     string         `json:"proto"`     // tcp, udp, icmp, any
	SrcIP     string         `json:"src_ip"`
	DestIP    string         `json:"dest_ip"`
	DestPort  string         `json:"dest_port"`
	Target    string         `json:"target" gorm:"not null"` // ACCEPT, REJECT, DROP
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	Position  int            `json:"position" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
