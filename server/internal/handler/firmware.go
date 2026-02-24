package handler

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

const firmwareDir = "./firmware_store"

type FirmwareHandler struct {
	DB   *gorm.DB
	MQTT mqtt.Client
}

func (h *FirmwareHandler) List(c *gin.Context) {
	var firmwares []model.Firmware
	query := h.DB
	if target := c.Query("target"); target != "" {
		query = query.Where("target = ?", target)
	}
	query.Order("created_at DESC").Find(&firmwares)
	c.JSON(http.StatusOK, firmwares)
}

func (h *FirmwareHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	version := c.PostForm("version")
	target := c.PostForm("target")
	changelog := c.PostForm("changelog")
	if version == "" || target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version and target are required"})
		return
	}

	os.MkdirAll(firmwareDir, 0755)
	savePath := filepath.Join(firmwareDir, file.Filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	hash, _ := hashFile(savePath)

	fw := model.Firmware{
		Version:     version,
		Target:      target,
		Filename:    file.Filename,
		FileSize:    file.Size,
		SHA256:      hash,
		DownloadURL: fmt.Sprintf("/api/v1/firmware/download/%s", file.Filename),
		Changelog:   changelog,
	}
	if err := h.DB.Create(&fw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(h.DB, c, "upload", "firmware", fmt.Sprintf("uploaded firmware %s v%s", file.Filename, version))
	c.JSON(http.StatusCreated, fw)
}

func (h *FirmwareHandler) Download(c *gin.Context) {
	filename := c.Param("filename")
	filePath := filepath.Join(firmwareDir, filepath.Base(filename))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	c.File(filePath)
}

func (h *FirmwareHandler) Delete(c *gin.Context) {
	var fw model.Firmware
	if err := h.DB.First(&fw, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "firmware not found"})
		return
	}

	os.Remove(filepath.Join(firmwareDir, fw.Filename))
	h.DB.Delete(&fw)
	writeAudit(h.DB, c, "delete", "firmware", fmt.Sprintf("deleted firmware %s (id=%d)", fw.Filename, fw.ID))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *FirmwareHandler) MarkStable(c *gin.Context) {
	var fw model.Firmware
	if err := h.DB.First(&fw, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "firmware not found"})
		return
	}
	h.DB.Model(&fw).Update("is_stable", true)
	c.JSON(http.StatusOK, gin.H{"message": "marked as stable"})
}

// PushUpgrade sends firmware upgrade command to a device.
func (h *FirmwareHandler) PushUpgrade(c *gin.Context) {
	var req struct {
		DeviceID   uint `json:"device_id" binding:"required"`
		FirmwareID uint `json:"firmware_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var device model.Device
	if err := h.DB.First(&device, req.DeviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	var fw model.Firmware
	if err := h.DB.First(&fw, req.FirmwareID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "firmware not found"})
		return
	}

	now := time.Now()
	upgrade := model.FirmwareUpgrade{
		DeviceID:   device.ID,
		FirmwareID: fw.ID,
		Status:     "pending",
		StartedAt:  &now,
	}
	h.DB.Create(&upgrade)

	if h.MQTT != nil && h.MQTT.IsConnected() {
		topic := fmt.Sprintf("nexusgate/devices/%s/command", device.MAC)
		payload := fmt.Sprintf(`{"action":"upgrade","url":"%s","sha256":"%s","version":"%s"}`,
			fw.DownloadURL, fw.SHA256, fw.Version)
		h.MQTT.Publish(topic, 1, false, payload)
	}

	writeAudit(h.DB, c, "upgrade", "firmware", fmt.Sprintf("pushed firmware v%s to device %s", fw.Version, device.Name))
	c.JSON(http.StatusOK, gin.H{"message": "upgrade initiated", "upgrade_id": upgrade.ID})
}

// BatchUpgrade pushes firmware to multiple devices matching a target.
func (h *FirmwareHandler) BatchUpgrade(c *gin.Context) {
	var req struct {
		FirmwareID uint   `json:"firmware_id" binding:"required"`
		Group      string `json:"group"`
		Model      string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var fw model.Firmware
	if err := h.DB.First(&fw, req.FirmwareID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "firmware not found"})
		return
	}

	var devices []model.Device
	query := h.DB.Where("status = ?", model.StatusOnline)
	if req.Group != "" {
		query = query.Where("\"group\" = ?", req.Group)
	}
	if req.Model != "" {
		query = query.Where("model LIKE ?", "%"+req.Model+"%")
	}
	query.Find(&devices)

	now := time.Now()
	count := 0
	for _, device := range devices {
		upgrade := model.FirmwareUpgrade{
			DeviceID:   device.ID,
			FirmwareID: fw.ID,
			Status:     "pending",
			StartedAt:  &now,
		}
		h.DB.Create(&upgrade)

		if h.MQTT != nil && h.MQTT.IsConnected() {
			topic := fmt.Sprintf("nexusgate/devices/%s/command", device.MAC)
			payload := fmt.Sprintf(`{"action":"upgrade","url":"%s","sha256":"%s","version":"%s"}`,
				fw.DownloadURL, fw.SHA256, fw.Version)
			h.MQTT.Publish(topic, 1, false, payload)
		}
		count++
	}

	writeAudit(h.DB, c, "batch_upgrade", "firmware", fmt.Sprintf("batch pushed firmware v%s to %d devices", fw.Version, count))
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("upgrade pushed to %d devices", count)})
}

func (h *FirmwareHandler) UpgradeHistory(c *gin.Context) {
	var upgrades []model.FirmwareUpgrade
	query := h.DB
	if deviceID := c.Query("device_id"); deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	query.Order("created_at DESC").Limit(100).Find(&upgrades)
	c.JSON(http.StatusOK, upgrades)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
