package model

import "time"

type AlertSeverity string

const (
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

type Alert struct {
	ID         uint          `json:"id" gorm:"primaryKey"`
	DeviceID   uint          `json:"device_id" gorm:"index;not null"`
	DeviceName string        `json:"device_name"`
	Metric     string        `json:"metric" gorm:"not null"` // cpu, memory, conntrack
	Value      float64       `json:"value"`
	Threshold  float64       `json:"threshold"`
	Severity   AlertSeverity `json:"severity" gorm:"default:warning"`
	Resolved   bool          `json:"resolved" gorm:"default:false;index"`
	CreatedAt  time.Time     `json:"created_at"`
	ResolvedAt *time.Time    `json:"resolved_at"`
}
