package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

// configEnvelope wraps UCI config content with an ID so the agent can ACK.
type configEnvelope struct {
	ConfigID uint   `json:"config_id"`
	Content  string `json:"content"`
}

// publishConfig sends a config envelope via MQTT and returns the JSON payload.
func publishConfig(mqttClient mqtt.Client, mac string, configID uint, content string) {
	if mqttClient == nil || !mqttClient.IsConnected() {
		return
	}
	envelope := configEnvelope{ConfigID: configID, Content: content}
	payload, _ := json.Marshal(envelope)
	topic := fmt.Sprintf("nexusgate/devices/%s/config", mac)
	mqttClient.Publish(topic, 1, false, payload)
}

type ConfigHandler struct {
	DB   *gorm.DB
	MQTT mqtt.Client
}

func (h *ConfigHandler) ListTemplates(c *gin.Context) {
	var templates []model.ConfigTemplate
	query := h.DB

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	query.Find(&templates)
	c.JSON(http.StatusOK, templates)
}

func (h *ConfigHandler) CreateTemplate(c *gin.Context) {
	var tpl model.ConfigTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&tpl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(h.DB, c, "create", "template", fmt.Sprintf("created template %s (id=%d)", tpl.Name, tpl.ID))
	c.JSON(http.StatusCreated, tpl)
}

func (h *ConfigHandler) UpdateTemplate(c *gin.Context) {
	var tpl model.ConfigTemplate
	if err := h.DB.First(&tpl, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	var req model.ConfigTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tpl.Name = req.Name
	tpl.Description = req.Description
	tpl.Category = req.Category
	tpl.Content = req.Content
	tpl.Version++

	h.DB.Save(&tpl)
	writeAudit(h.DB, c, "update", "template", fmt.Sprintf("updated template %s (id=%d) to v%d", tpl.Name, tpl.ID, tpl.Version))
	c.JSON(http.StatusOK, tpl)
}

func (h *ConfigHandler) DeleteTemplate(c *gin.Context) {
	if err := h.DB.Delete(&model.ConfigTemplate{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(h.DB, c, "delete", "template", fmt.Sprintf("deleted template id=%s", c.Param("id")))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) PushConfig(c *gin.Context) {
	var device model.Device
	if err := h.DB.First(&device, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	var req struct {
		TemplateID *uint  `json:"template_id"`
		Content    string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content := req.Content
	if req.TemplateID != nil {
		var tpl model.ConfigTemplate
		if err := h.DB.First(&tpl, *req.TemplateID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		content = tpl.Content
	}

	record := model.DeviceConfig{
		DeviceID:   device.ID,
		TemplateID: req.TemplateID,
		Content:    content,
		Status:     "pending",
	}
	h.DB.Create(&record)

	publishConfig(h.MQTT, device.MAC, record.ID, content)
	writeAudit(h.DB, c, "push", "config", fmt.Sprintf("pushed config to device %s (config_id=%d)", device.Name, record.ID))
	c.JSON(http.StatusOK, gin.H{"message": "config push initiated", "config_id": record.ID})
}

func (h *ConfigHandler) ConfigHistory(c *gin.Context) {
	var configs []model.DeviceConfig
	h.DB.Where("device_id = ?", c.Param("id")).
		Order("created_at DESC").Limit(50).
		Find(&configs)
	c.JSON(http.StatusOK, configs)
}
