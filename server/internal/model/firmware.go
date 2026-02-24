package model

import (
	"time"

	"gorm.io/gorm"
)

type Firmware struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Version     string         `json:"version" gorm:"not null"`
	Target      string         `json:"target" gorm:"not null"`   // e.g. x86-64, nanopi-r4s
	Filename    string         `json:"filename" gorm:"not null"`
	FileSize    int64          `json:"file_size"`
	SHA256      string         `json:"sha256"`
	DownloadURL string         `json:"download_url"`
	Changelog   string         `json:"changelog" gorm:"type:text"`
	IsStable    bool           `json:"is_stable" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type FirmwareUpgrade struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	DeviceID   uint       `json:"device_id" gorm:"index;not null"`
	FirmwareID uint       `json:"firmware_id" gorm:"not null"`
	Status     string     `json:"status" gorm:"default:pending"` // pending, downloading, upgrading, success, failed
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	ErrorMsg   string     `json:"error_msg"`
	CreatedAt  time.Time  `json:"created_at"`
}
