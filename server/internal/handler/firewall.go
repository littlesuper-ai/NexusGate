package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

type FirewallHandler struct {
	DB   *gorm.DB
	MQTT mqtt.Client
}

// --- Zones ---

func (h *FirewallHandler) ListZones(c *gin.Context) {
	deviceID := c.Query("device_id")
	var zones []model.FirewallZone
	query := h.DB
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	query.Order("id").Find(&zones)
	c.JSON(http.StatusOK, zones)
}

func (h *FirewallHandler) CreateZone(c *gin.Context) {
	var zone model.FirewallZone
	if err := c.ShouldBindJSON(&zone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&zone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(h.DB, c, "create", "firewall_zone", fmt.Sprintf("created firewall zone %s (id=%d)", zone.Name, zone.ID))
	c.JSON(http.StatusCreated, zone)
}

func (h *FirewallHandler) UpdateZone(c *gin.Context) {
	var zone model.FirewallZone
	if err := h.DB.First(&zone, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
		return
	}
	if err := c.ShouldBindJSON(&zone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&zone)
	writeAudit(h.DB, c, "update", "firewall_zone", fmt.Sprintf("updated firewall zone %s (id=%d)", zone.Name, zone.ID))
	c.JSON(http.StatusOK, zone)
}

func (h *FirewallHandler) DeleteZone(c *gin.Context) {
	if err := h.DB.Delete(&model.FirewallZone{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(h.DB, c, "delete", "firewall_zone", fmt.Sprintf("deleted firewall zone id=%s", c.Param("id")))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Rules ---

func (h *FirewallHandler) ListRules(c *gin.Context) {
	deviceID := c.Query("device_id")
	var rules []model.FirewallRule
	query := h.DB
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	query.Order("position, id").Find(&rules)
	c.JSON(http.StatusOK, rules)
}

func (h *FirewallHandler) CreateRule(c *gin.Context) {
	var rule model.FirewallRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(h.DB, c, "create", "firewall_rule", fmt.Sprintf("created firewall rule %s (id=%d)", rule.Name, rule.ID))
	c.JSON(http.StatusCreated, rule)
}

func (h *FirewallHandler) UpdateRule(c *gin.Context) {
	var rule model.FirewallRule
	if err := h.DB.First(&rule, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&rule)
	writeAudit(h.DB, c, "update", "firewall_rule", fmt.Sprintf("updated firewall rule %s (id=%d)", rule.Name, rule.ID))
	c.JSON(http.StatusOK, rule)
}

func (h *FirewallHandler) DeleteRule(c *gin.Context) {
	if err := h.DB.Delete(&model.FirewallRule{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(h.DB, c, "delete", "firewall_rule", fmt.Sprintf("deleted firewall rule id=%s", c.Param("id")))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ApplyFirewall generates UCI firewall config and pushes it to the device.
func (h *FirewallHandler) ApplyFirewall(c *gin.Context) {
	deviceID := c.Param("device_id")

	var device model.Device
	if err := h.DB.First(&device, deviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	var zones []model.FirewallZone
	h.DB.Where("device_id = ?", deviceID).Find(&zones)

	var rules []model.FirewallRule
	h.DB.Where("device_id = ? AND enabled = true", deviceID).Order("position, id").Find(&rules)

	uci := generateFirewallUCI(zones, rules)

	record := model.DeviceConfig{DeviceID: device.ID, Content: uci, Status: "pending"}
	h.DB.Create(&record)
	publishConfig(h.MQTT, device.MAC, record.ID, uci)

	writeAudit(h.DB, c, "apply", "firewall", fmt.Sprintf("applied firewall config to device %s", device.Name))
	c.JSON(http.StatusOK, gin.H{"message": "firewall config pushed", "config_id": record.ID})
}

func generateFirewallUCI(zones []model.FirewallZone, rules []model.FirewallRule) string {
	var b strings.Builder
	b.WriteString("package firewall\n\n")

	b.WriteString("config defaults\n")
	b.WriteString("\toption syn_flood '1'\n")
	b.WriteString("\toption input 'REJECT'\n")
	b.WriteString("\toption output 'ACCEPT'\n")
	b.WriteString("\toption forward 'REJECT'\n\n")

	for _, z := range zones {
		b.WriteString(fmt.Sprintf("config zone '%s'\n", z.Name))
		b.WriteString(fmt.Sprintf("\toption name '%s'\n", z.Name))
		b.WriteString(fmt.Sprintf("\toption input '%s'\n", z.Input))
		b.WriteString(fmt.Sprintf("\toption output '%s'\n", z.Output))
		b.WriteString(fmt.Sprintf("\toption forward '%s'\n", z.Forward))
		if z.Masq {
			b.WriteString("\toption masq '1'\n")
			b.WriteString("\toption mtu_fix '1'\n")
		}
		for _, net := range strings.Split(z.Networks, ",") {
			net = strings.TrimSpace(net)
			if net != "" {
				b.WriteString(fmt.Sprintf("\tlist network '%s'\n", net))
			}
		}
		b.WriteString("\n")
	}

	for _, r := range rules {
		b.WriteString(fmt.Sprintf("config rule '%s'\n", r.Name))
		b.WriteString(fmt.Sprintf("\toption name '%s'\n", r.Name))
		if r.Src != "" {
			b.WriteString(fmt.Sprintf("\toption src '%s'\n", r.Src))
		}
		if r.Dest != "" {
			b.WriteString(fmt.Sprintf("\toption dest '%s'\n", r.Dest))
		}
		if r.Proto != "" && r.Proto != "any" {
			b.WriteString(fmt.Sprintf("\toption proto '%s'\n", r.Proto))
		}
		if r.SrcIP != "" {
			b.WriteString(fmt.Sprintf("\toption src_ip '%s'\n", r.SrcIP))
		}
		if r.DestIP != "" {
			b.WriteString(fmt.Sprintf("\toption dest_ip '%s'\n", r.DestIP))
		}
		if r.DestPort != "" {
			b.WriteString(fmt.Sprintf("\toption dest_port '%s'\n", r.DestPort))
		}
		b.WriteString(fmt.Sprintf("\toption target '%s'\n", r.Target))
		b.WriteString("\n")
	}

	return b.String()
}
