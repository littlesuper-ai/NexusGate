package mqtt

import (
	"encoding/json"
	"log"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/config"
	"github.com/nexusgate/nexusgate/internal/model"
	"github.com/nexusgate/nexusgate/internal/ws"
	"gorm.io/gorm"
)

func NewClient(cfg *config.Config) (pahomqtt.Client, error) {
	opts := pahomqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID("nexusgate-server").
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second)

	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return nil, token.Error()
	}
	if token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

// SubscribeDeviceStatus listens for heartbeat messages from agents and updates device status.
// It also broadcasts status updates to connected WebSocket clients via the hub.
func SubscribeDeviceStatus(client pahomqtt.Client, db *gorm.DB, hub *ws.Hub) {
	client.Subscribe("nexusgate/devices/+/status", 1, func(_ pahomqtt.Client, msg pahomqtt.Message) {
		var payload struct {
			MAC        string  `json:"mac"`
			CPUUsage   float64 `json:"cpu_usage"`
			MemUsage   float64 `json:"mem_usage"`
			MemTotal   int64   `json:"mem_total"`
			MemFree    int64   `json:"mem_free"`
			RxBytes    int64   `json:"rx_bytes"`
			TxBytes    int64   `json:"tx_bytes"`
			Conntrack  int     `json:"conntrack"`
			UptimeSecs int64   `json:"uptime_secs"`
			LoadAvg    string  `json:"load_avg"`
		}
		if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
			log.Printf("invalid status payload: %v", err)
			return
		}

		now := time.Now()
		db.Model(&model.Device{}).Where("mac = ?", payload.MAC).Updates(map[string]any{
			"status":       model.StatusOnline,
			"cpu_usage":    payload.CPUUsage,
			"mem_usage":    payload.MemUsage,
			"uptime_secs":  payload.UptimeSecs,
			"last_seen_at": &now,
		})

		// Look up device ID for metrics record
		var device model.Device
		var deviceID uint
		if err := db.Select("id").Where("mac = ?", payload.MAC).First(&device).Error; err == nil {
			deviceID = device.ID
		}

		db.Create(&model.DeviceMetrics{
			DeviceID:    deviceID,
			CPUUsage:    payload.CPUUsage,
			MemUsage:    payload.MemUsage,
			MemTotal:    payload.MemTotal,
			MemFree:     payload.MemFree,
			RxBytes:     payload.RxBytes,
			TxBytes:     payload.TxBytes,
			Conntrack:   payload.Conntrack,
			UptimeSecs:  payload.UptimeSecs,
			LoadAvg:     payload.LoadAvg,
			CollectedAt: now,
		})

		// Broadcast to WebSocket clients
		if hub != nil {
			hub.Broadcast("device_status", map[string]any{
				"mac":         payload.MAC,
				"device_id":   deviceID,
				"cpu_usage":   payload.CPUUsage,
				"mem_usage":   payload.MemUsage,
				"rx_bytes":    payload.RxBytes,
				"tx_bytes":    payload.TxBytes,
				"conntrack":   payload.Conntrack,
				"uptime_secs": payload.UptimeSecs,
				"load_avg":    payload.LoadAvg,
				"status":      "online",
			})
		}
	})
}
