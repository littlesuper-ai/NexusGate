package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

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
	c.JSON(http.StatusOK, tpl)
}

func (h *ConfigHandler) DeleteTemplate(c *gin.Context) {
	if err := h.DB.Delete(&model.ConfigTemplate{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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

	// Push config via MQTT
	if h.MQTT != nil && h.MQTT.IsConnected() {
		topic := fmt.Sprintf("nexusgate/devices/%s/config", device.MAC)
		h.MQTT.Publish(topic, 1, false, content)
	}

	c.JSON(http.StatusOK, gin.H{"message": "config push initiated", "config_id": record.ID})
}

func (h *ConfigHandler) ConfigHistory(c *gin.Context) {
	var configs []model.DeviceConfig
	h.DB.Where("device_id = ?", c.Param("id")).
		Order("created_at DESC").Limit(50).
		Find(&configs)
	c.JSON(http.StatusOK, configs)
}
