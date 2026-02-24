package jobs

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nexusgate/nexusgate/internal/model"
	"github.com/nexusgate/nexusgate/internal/ws"
	"gorm.io/gorm"
)

type alertThresholds struct {
	CPU       float64
	Memory    float64
	Conntrack int
}

func getAlertThresholds(db *gorm.DB) alertThresholds {
	t := alertThresholds{CPU: 90, Memory: 90, Conntrack: 50000}
	readFloat := func(key string, target *float64) {
		var s model.SystemSetting
		if err := db.Where("\"key\" = ?", key).First(&s).Error; err == nil {
			if v, err := strconv.ParseFloat(s.Value, 64); err == nil && v > 0 {
				*target = v
			}
		}
	}
	readFloat("alert_cpu_threshold", &t.CPU)
	readFloat("alert_mem_threshold", &t.Memory)
	var s model.SystemSetting
	if err := db.Where("\"key\" = ?", "alert_conntrack_threshold").First(&s).Error; err == nil {
		if v, err := strconv.Atoi(s.Value); err == nil && v > 0 {
			t.Conntrack = v
		}
	}
	return t
}

// EvaluateDeviceAlerts checks a single heartbeat against thresholds and creates alerts.
// Called from the MQTT handler on each heartbeat.
func EvaluateDeviceAlerts(db *gorm.DB, hub *ws.Hub, deviceID uint, deviceName string, cpuUsage, memUsage float64, conntrack int) {
	t := getAlertThresholds(db)
	now := time.Now()

	check := func(metric string, value, threshold float64) {
		if value < threshold {
			// Auto-resolve if previously alerting
			db.Model(&model.Alert{}).
				Where("device_id = ? AND metric = ? AND resolved = false", deviceID, metric).
				Updates(map[string]any{"resolved": true, "resolved_at": &now})
			return
		}
		// Check if there's already an unresolved alert for this device+metric
		var existing model.Alert
		if err := db.Where("device_id = ? AND metric = ? AND resolved = false", deviceID, metric).
			First(&existing).Error; err == nil {
			// Update value on existing alert
			db.Model(&existing).Update("value", value)
			return
		}
		// Create new alert
		alert := model.Alert{
			DeviceID:   deviceID,
			DeviceName: deviceName,
			Metric:     metric,
			Value:      value,
			Threshold:  threshold,
			Severity:   model.SeverityWarning,
		}
		if value > threshold*1.2 {
			alert.Severity = model.SeverityCritical
		}
		db.Create(&alert)
		log.Printf("ALERT: device=%s metric=%s value=%.1f threshold=%.1f", deviceName, metric, value, threshold)

		// Broadcast to WebSocket
		if hub != nil {
			hub.Broadcast("alert", map[string]any{
				"id":          alert.ID,
				"device_id":   deviceID,
				"device_name": deviceName,
				"metric":      metric,
				"value":       value,
				"threshold":   threshold,
				"severity":    alert.Severity,
			})
		}

		// Dispatch notification
		dispatchNotification(db, alert)
	}

	check("cpu", cpuUsage, t.CPU)
	check("memory", memUsage, t.Memory)
	check("conntrack", float64(conntrack), float64(t.Conntrack))
}

func dispatchNotification(db *gorm.DB, alert model.Alert) {
	var methodSetting model.SystemSetting
	if err := db.Where("\"key\" = ?", "alert_notify_method").First(&methodSetting).Error; err != nil {
		return // No notification method configured
	}

	switch methodSetting.Value {
	case "webhook":
		var urlSetting model.SystemSetting
		if err := db.Where("\"key\" = ?", "alert_webhook_url").First(&urlSetting).Error; err != nil || urlSetting.Value == "" {
			log.Println("alert webhook URL not configured")
			return
		}
		go sendWebhook(urlSetting.Value, alert)
	case "log":
		log.Printf("ALERT NOTIFICATION [%s]: device=%s metric=%s value=%.1f threshold=%.1f",
			alert.Severity, alert.DeviceName, alert.Metric, alert.Value, alert.Threshold)
	}
}

func sendWebhook(url string, alert model.Alert) {
	payload, _ := json.Marshal(map[string]any{
		"device_name": alert.DeviceName,
		"device_id":   alert.DeviceID,
		"metric":      alert.Metric,
		"value":       alert.Value,
		"threshold":   alert.Threshold,
		"severity":    alert.Severity,
		"time":        alert.CreatedAt.Format(time.RFC3339),
	})

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("webhook send failed: %v", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("webhook returned status %d", resp.StatusCode)
	}
}
