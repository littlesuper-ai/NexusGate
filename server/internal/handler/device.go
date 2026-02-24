package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

type DeviceHandler struct {
	DB   *gorm.DB
	MQTT mqtt.Client
}

// Register handles device self-registration (called by nexusgate-agent on first boot).
func (h *DeviceHandler) Register(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		MAC       string `json:"mac" binding:"required"`
		IPAddress string `json:"ip_address"`
		Model     string `json:"model"`
		Firmware  string `json:"firmware"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := model.Device{
		Name:      req.Name,
		MAC:       req.MAC,
		IPAddress: req.IPAddress,
		Model:     req.Model,
		Firmware:  req.Firmware,
		Status:    model.StatusOnline,
	}

	// Upsert: update if MAC exists, create otherwise
	result := h.DB.Where("mac = ?", req.MAC).FirstOrCreate(&device)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		now := time.Now()
		h.DB.Model(&device).Updates(map[string]any{
			"ip_address":   req.IPAddress,
			"firmware":     req.Firmware,
			"status":       model.StatusOnline,
			"last_seen_at": &now,
		})
	}

	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) List(c *gin.Context) {
	var devices []model.Device
	query := h.DB

	if group := c.Query("group"); group != "" {
		query = query.Where("\"group\" = ?", group)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, devices)
}

func (h *DeviceHandler) Get(c *gin.Context) {
	var device model.Device
	if err := h.DB.First(&device, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}
	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) Update(c *gin.Context) {
	var device model.Device
	if err := h.DB.First(&device, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Group string `json:"group"`
		Tags  string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.DB.Model(&device).Updates(map[string]any{
		"name":  req.Name,
		"group": req.Group,
		"tags":  req.Tags,
	})
	writeAudit(h.DB, c, "update", "device", fmt.Sprintf("updated device %s (id=%d)", device.Name, device.ID))
	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) Delete(c *gin.Context) {
	if err := h.DB.Delete(&model.Device{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(h.DB, c, "delete", "device", fmt.Sprintf("deleted device id=%s", c.Param("id")))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *DeviceHandler) Reboot(c *gin.Context) {
	var device model.Device
	if err := h.DB.First(&device, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	if h.MQTT != nil && h.MQTT.IsConnected() {
		topic := fmt.Sprintf("nexusgate/devices/%s/command", device.MAC)
		h.MQTT.Publish(topic, 1, false, `{"action":"reboot"}`)
	}

	writeAudit(h.DB, c, "reboot", "device", fmt.Sprintf("rebooted device %s (id=%d)", device.Name, device.ID))
	c.JSON(http.StatusOK, gin.H{"message": "reboot command sent"})
}

func (h *DeviceHandler) Metrics(c *gin.Context) {
	var metrics []model.DeviceMetrics
	query := h.DB.Where("device_id = ?", c.Param("id"))

	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			query = query.Where("collected_at >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			query = query.Where("collected_at <= ?", t)
		}
	}
	if hours := c.Query("hours"); hours != "" {
		var n int
		if _, err := fmt.Sscanf(hours, "%d", &n); err == nil && n > 0 {
			query = query.Where("collected_at >= ?", time.Now().Add(-time.Duration(n)*time.Hour))
		}
	}

	if err := query.Order("collected_at DESC").Limit(500).
		Find(&metrics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *DeviceHandler) DashboardSummary(c *gin.Context) {
	var total, online, offline int64

	h.DB.Model(&model.Device{}).Count(&total)
	h.DB.Model(&model.Device{}).Where("status = ?", model.StatusOnline).Count(&online)
	h.DB.Model(&model.Device{}).Where("status = ?", model.StatusOffline).Count(&offline)

	c.JSON(http.StatusOK, gin.H{
		"total_devices":   total,
		"online_devices":  online,
		"offline_devices": offline,
	})
}
