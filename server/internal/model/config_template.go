package model

import (
	"time"

	"gorm.io/gorm"
)

type ConfigTemplate struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	Category    string         `json:"category" gorm:"index"` // network, firewall, vpn, qos, etc.
	Content     string         `json:"content" gorm:"type:text;not null"`
	Version     int            `json:"version" gorm:"default:1"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type DeviceConfig struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DeviceID   uint      `json:"device_id" gorm:"index;not null"`
	TemplateID *uint     `json:"template_id"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	Version    int       `json:"version" gorm:"default:1"`
	AppliedAt  *time.Time `json:"applied_at"`
	Status     string    `json:"status" gorm:"default:pending"` // pending, applied, failed
	CreatedAt  time.Time `json:"created_at"`
}
