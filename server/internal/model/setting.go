package model

import "time"

type SystemSetting struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null"`
	Value     string    `json:"value" gorm:"type:text"`
	Category  string    `json:"category" gorm:"index;default:general"`
	UpdatedAt time.Time `json:"updated_at"`
}
