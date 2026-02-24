package jobs

import (
	"log"
	"strconv"
	"time"

	"github.com/nexusgate/nexusgate/internal/model"
	"github.com/nexusgate/nexusgate/internal/ws"
	"gorm.io/gorm"
)

const defaultOfflineThreshold = 120 // seconds

// StartOfflineDetector runs a periodic check that marks devices as offline
// when their last_seen_at exceeds the configured threshold. It also broadcasts
// status changes to WebSocket clients.
func StartOfflineDetector(db *gorm.DB, hub *ws.Hub) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			checkOfflineDevices(db, hub)
		}
	}()
	log.Println("offline detector started (interval: 30s)")
}

func checkOfflineDevices(db *gorm.DB, hub *ws.Hub) {
	threshold := defaultOfflineThreshold

	// Try to read the threshold from system settings
	var setting model.SystemSetting
	if err := db.Where("`key` = ?", "offline_threshold").First(&setting).Error; err == nil {
		if v, err := strconv.Atoi(setting.Value); err == nil && v > 0 {
			threshold = v
		}
	}

	cutoff := time.Now().Add(-time.Duration(threshold) * time.Second)

	var staleDevices []model.Device
	db.Where("status = ? AND last_seen_at < ?", model.StatusOnline, cutoff).Find(&staleDevices)

	if len(staleDevices) == 0 {
		return
	}

	ids := make([]uint, len(staleDevices))
	for i, d := range staleDevices {
		ids[i] = d.ID
	}

	db.Model(&model.Device{}).Where("id IN ?", ids).Update("status", model.StatusOffline)

	log.Printf("marked %d device(s) as offline (threshold: %ds)", len(staleDevices), threshold)

	// Broadcast offline events
	if hub != nil {
		for _, d := range staleDevices {
			hub.Broadcast("device_status", map[string]any{
				"mac":       d.MAC,
				"device_id": d.ID,
				"status":    "offline",
			})
		}
	}
}

// StartMetricsCleanup runs a daily cleanup of old device metrics.
func StartMetricsCleanup(db *gorm.DB) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupOldMetrics(db)
		}
	}()
	log.Println("metrics cleanup job started (interval: 24h)")
}

func cleanupOldMetrics(db *gorm.DB) {
	retentionDays := 30

	var setting model.SystemSetting
	if err := db.Where("`key` = ?", "metrics_retention_days").First(&setting).Error; err == nil {
		if v, err := strconv.Atoi(setting.Value); err == nil && v > 0 {
			retentionDays = v
		}
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := db.Where("collected_at < ?", cutoff).Delete(&model.DeviceMetrics{})
	if result.RowsAffected > 0 {
		log.Printf("cleaned up %d old metric records (retention: %d days)", result.RowsAffected, retentionDays)
	}
}
