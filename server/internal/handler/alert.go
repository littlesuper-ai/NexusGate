package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

type AlertHandler struct {
	DB *gorm.DB
}

func (h *AlertHandler) List(c *gin.Context) {
	var alerts []model.Alert
	query := h.DB

	if deviceID := c.Query("device_id"); deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if resolved := c.Query("resolved"); resolved == "false" {
		query = query.Where("resolved = false")
	} else if resolved == "true" {
		query = query.Where("resolved = true")
	}
	if severity := c.Query("severity"); severity != "" {
		switch severity {
		case string(model.SeverityWarning), string(model.SeverityCritical):
			query = query.Where("severity = ?", severity)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid severity value"})
			return
		}
	}

	// Pagination
	page := 1
	pageSize := 50
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 200 {
			pageSize = v
		}
	}

	var total int64
	query.Count(&total)
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&alerts)
	c.JSON(http.StatusOK, gin.H{"data": alerts, "total": total, "page": page, "page_size": pageSize})
}

func (h *AlertHandler) Resolve(c *gin.Context) {
	var alert model.Alert
	if err := h.DB.First(&alert, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}
	now := time.Now()
	h.DB.Model(&alert).Updates(map[string]any{"resolved": true, "resolved_at": &now})
	c.JSON(http.StatusOK, gin.H{"message": "resolved"})
}

func (h *AlertHandler) Summary(c *gin.Context) {
	var total, unresolved, warning, critical int64
	h.DB.Model(&model.Alert{}).Count(&total)
	h.DB.Model(&model.Alert{}).Where("resolved = false").Count(&unresolved)
	h.DB.Model(&model.Alert{}).Where("resolved = false AND severity = ?", model.SeverityWarning).Count(&warning)
	h.DB.Model(&model.Alert{}).Where("resolved = false AND severity = ?", model.SeverityCritical).Count(&critical)

	c.JSON(http.StatusOK, gin.H{
		"total":      total,
		"unresolved": unresolved,
		"warning":    warning,
		"critical":   critical,
	})
}
