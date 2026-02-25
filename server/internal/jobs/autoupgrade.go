package jobs

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

// StartAutoUpgradeChecker runs a periodic job that checks for online devices
// running firmware older than the latest stable version and pushes upgrades.
// Controlled by the "firmware_auto_upgrade" system setting (value "true" to enable).
func StartAutoUpgradeChecker(db *gorm.DB, mqttClient mqtt.Client) {
	go func() {
		// Check every hour
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			runAutoUpgrade(db, mqttClient)
		}
	}()
	log.Println("firmware auto-upgrade checker started (interval: 1h)")
}

func runAutoUpgrade(db *gorm.DB, mqttClient mqtt.Client) {
	// Check if auto-upgrade is enabled
	var setting model.SystemSetting
	if err := db.Where("\"key\" = ?", "firmware_auto_upgrade").First(&setting).Error; err != nil {
		return // Setting not found, skip
	}
	if setting.Value != "true" {
		return
	}

	// Find stable firmware grouped by target (latest per target)
	var stableFirmwares []model.Firmware
	db.Where("is_stable = true").Order("created_at DESC").Find(&stableFirmwares)

	// Build a map of target -> latest stable firmware
	latestByTarget := make(map[string]model.Firmware)
	for _, fw := range stableFirmwares {
		if _, exists := latestByTarget[fw.Target]; !exists {
			latestByTarget[fw.Target] = fw
		}
	}

	if len(latestByTarget) == 0 {
		return
	}

	// Find online devices that are running older firmware
	var devices []model.Device
	db.Where("status = ?", model.StatusOnline).Find(&devices)

	if mqttClient == nil || !mqttClient.IsConnected() {
		return
	}

	upgradedCount := 0
	now := time.Now()

	for _, device := range devices {
		// Match device model to firmware target
		fw, exists := latestByTarget[device.Model]
		if !exists {
			continue
		}

		// Skip if device is already on this firmware version
		if device.Firmware == fw.Version {
			continue
		}

		// Skip if there's already a pending/in-progress upgrade for this device
		var pendingCount int64
		db.Model(&model.FirmwareUpgrade{}).
			Where("device_id = ? AND status IN ?", device.ID, []string{"pending", "downloading", "upgrading"}).
			Count(&pendingCount)
		if pendingCount > 0 {
			continue
		}

		// Create upgrade record and push command
		upgrade := model.FirmwareUpgrade{
			DeviceID:   device.ID,
			FirmwareID: fw.ID,
			Status:     "pending",
			StartedAt:  &now,
		}
		db.Create(&upgrade)

		topic := fmt.Sprintf("nexusgate/devices/%s/command", device.MAC)
		payload := fmt.Sprintf(`{"action":"upgrade","url":"%s","sha256":"%s","version":"%s"}`,
			fw.DownloadURL, fw.SHA256, fw.Version)
		mqttClient.Publish(topic, 1, false, payload)
		upgradedCount++
	}

	if upgradedCount > 0 {
		log.Printf("auto-upgrade: pushed firmware upgrades to %d device(s)", upgradedCount)
	}
}
