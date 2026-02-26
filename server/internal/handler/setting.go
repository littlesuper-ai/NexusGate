package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettingHandler struct {
	DB *gorm.DB
}

func (h *SettingHandler) List(c *gin.Context) {
	var items []model.SystemSetting
	query := h.DB
	if cat := c.Query("category"); cat != "" {
		query = query.Where("category = ?", cat)
	}
	if err := query.Order("category, key").Limit(500).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query settings"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *SettingHandler) Get(c *gin.Context) {
	key := c.Param("key")
	var item model.SystemSetting
	if err := h.DB.Where("key = ?", key).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *SettingHandler) Upsert(c *gin.Context) {
	var req struct {
		Key      string `json:"key" binding:"required"`
		Value    string `json:"value"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Category == "" {
		req.Category = "general"
	}
	item := model.SystemSetting{Key: req.Key, Value: req.Value, Category: req.Category}
	if err := h.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "category", "updated_at"}),
	}).Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
		return
	}
	writeAudit(h.DB, c, "upsert", "setting", fmt.Sprintf("set setting %s=%s", req.Key, req.Value))
	c.JSON(http.StatusOK, item)
}

func (h *SettingHandler) BatchUpsert(c *gin.Context) {
	var items []struct {
		Key      string `json:"key" binding:"required"`
		Value    string `json:"value"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, item := range items {
		if item.Category == "" {
			item.Category = "general"
		}
		s := model.SystemSetting{Key: item.Key, Value: item.Value, Category: item.Category}
		if err := h.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "category", "updated_at"}),
		}).Create(&s).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save setting %s", item.Key)})
			return
		}
	}
	writeAudit(h.DB, c, "batch_upsert", "setting", fmt.Sprintf("batch updated %d settings", len(items)))
	c.JSON(http.StatusOK, gin.H{"message": "saved"})
}

func (h *SettingHandler) Delete(c *gin.Context) {
	result := h.DB.Where("key = ?", c.Param("key")).Delete(&model.SystemSetting{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}
	writeAudit(h.DB, c, "delete", "setting", fmt.Sprintf("deleted setting %s", c.Param("key")))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
