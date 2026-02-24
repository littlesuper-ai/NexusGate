package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

// mwanMember represents a member entry in the MWANPolicy.Members JSON.
type mwanMember struct {
	Iface  string `json:"iface"`
	Metric int    `json:"metric"`
	Weight int    `json:"weight"`
}

type NetworkHandler struct {
	DB   *gorm.DB
	MQTT mqtt.Client
}

// ==================== Multi-WAN ====================

func (h *NetworkHandler) ListWANInterfaces(c *gin.Context) {
	var items []model.WANInterface
	query := h.DB
	if did := c.Query("device_id"); did != "" {
		query = query.Where("device_id = ?", did)
	}
	query.Order("id").Find(&items)
	c.JSON(http.StatusOK, items)
}

func (h *NetworkHandler) CreateWANInterface(c *gin.Context) {
	var item model.WANInterface
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&item)
	c.JSON(http.StatusCreated, item)
}

func (h *NetworkHandler) UpdateWANInterface(c *gin.Context) {
	var item model.WANInterface
	if err := h.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wan interface not found"})
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

func (h *NetworkHandler) DeleteWANInterface(c *gin.Context) {
	h.DB.Delete(&model.WANInterface{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *NetworkHandler) ListMWANPolicies(c *gin.Context) {
	var items []model.MWANPolicy
	query := h.DB
	if did := c.Query("device_id"); did != "" {
		query = query.Where("device_id = ?", did)
	}
	query.Find(&items)
	c.JSON(http.StatusOK, items)
}

func (h *NetworkHandler) CreateMWANPolicy(c *gin.Context) {
	var item model.MWANPolicy
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&item)
	c.JSON(http.StatusCreated, item)
}

func (h *NetworkHandler) UpdateMWANPolicy(c *gin.Context) {
	var item model.MWANPolicy
	if err := h.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

func (h *NetworkHandler) DeleteMWANPolicy(c *gin.Context) {
	h.DB.Delete(&model.MWANPolicy{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *NetworkHandler) ListMWANRules(c *gin.Context) {
	var items []model.MWANRule
	query := h.DB
	if did := c.Query("device_id"); did != "" {
		query = query.Where("device_id = ?", did)
	}
	query.Order("position, id").Find(&items)
	c.JSON(http.StatusOK, items)
}

func (h *NetworkHandler) CreateMWANRule(c *gin.Context) {
	var item model.MWANRule
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&item)
	c.JSON(http.StatusCreated, item)
}

func (h *NetworkHandler) UpdateMWANRule(c *gin.Context) {
	var item model.MWANRule
	if err := h.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

func (h *NetworkHandler) DeleteMWANRule(c *gin.Context) {
	h.DB.Delete(&model.MWANRule{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *NetworkHandler) ApplyMWAN(c *gin.Context) {
	deviceID := c.Param("device_id")
	var device model.Device
	if err := h.DB.First(&device, deviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	var wans []model.WANInterface
	h.DB.Where("device_id = ? AND enabled = true", deviceID).Find(&wans)
	var policies []model.MWANPolicy
	h.DB.Where("device_id = ?", deviceID).Find(&policies)
	var rules []model.MWANRule
	h.DB.Where("device_id = ? AND enabled = true", deviceID).Order("position, id").Find(&rules)

	uci := generateMWANUCI(wans, policies, rules)

	if h.MQTT != nil && h.MQTT.IsConnected() {
		topic := fmt.Sprintf("nexusgate/devices/%s/config", device.MAC)
		h.MQTT.Publish(topic, 1, false, uci)
	}
	h.DB.Create(&model.DeviceConfig{DeviceID: device.ID, Content: uci, Status: "pending"})
	writeAudit(h.DB, c, "apply", "mwan", fmt.Sprintf("applied MWAN config to device %s", device.Name))
	c.JSON(http.StatusOK, gin.H{"message": "mwan3 config pushed", "uci": uci})
}

func generateMWANUCI(wans []model.WANInterface, policies []model.MWANPolicy, rules []model.MWANRule) string {
	var b strings.Builder
	for _, w := range wans {
		b.WriteString(fmt.Sprintf("config interface '%s'\n", w.Name))
		b.WriteString("\toption enabled '1'\n")
		for _, ip := range strings.Split(w.TrackIPs, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				b.WriteString(fmt.Sprintf("\tlist track_ip '%s'\n", ip))
			}
		}
		b.WriteString(fmt.Sprintf("\toption reliability '%d'\n", w.Reliability))
		b.WriteString(fmt.Sprintf("\toption count '3'\n"))
		b.WriteString(fmt.Sprintf("\toption timeout '3'\n"))
		b.WriteString(fmt.Sprintf("\toption interval '%d'\n", w.Interval))
		b.WriteString(fmt.Sprintf("\toption down '%d'\n", w.Down))
		b.WriteString(fmt.Sprintf("\toption up '%d'\n", w.Up))
		b.WriteString("\n")
	}
	for _, p := range policies {
		// Generate config member stanzas from the JSON members list
		var members []mwanMember
		if p.Members != "" {
			json.Unmarshal([]byte(p.Members), &members)
		}
		for _, m := range members {
			memberName := fmt.Sprintf("%s_m%d_w%d", m.Iface, m.Metric, m.Weight)
			b.WriteString(fmt.Sprintf("config member '%s'\n", memberName))
			b.WriteString(fmt.Sprintf("\toption interface '%s'\n", m.Iface))
			b.WriteString(fmt.Sprintf("\toption metric '%d'\n", m.Metric))
			b.WriteString(fmt.Sprintf("\toption weight '%d'\n", m.Weight))
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("config policy '%s'\n", p.Name))
		b.WriteString(fmt.Sprintf("\toption last_resort '%s'\n", p.LastResort))
		for _, m := range members {
			memberName := fmt.Sprintf("%s_m%d_w%d", m.Iface, m.Metric, m.Weight)
			b.WriteString(fmt.Sprintf("\tlist use_member '%s'\n", memberName))
		}
		b.WriteString("\n")
	}
	for _, r := range rules {
		b.WriteString(fmt.Sprintf("config rule '%s'\n", r.Name))
		if r.SrcIP != "" {
			b.WriteString(fmt.Sprintf("\toption src_ip '%s'\n", r.SrcIP))
		}
		if r.DestIP != "" {
			b.WriteString(fmt.Sprintf("\toption dest_ip '%s'\n", r.DestIP))
		}
		if r.Proto != "" && r.Proto != "all" {
			b.WriteString(fmt.Sprintf("\toption proto '%s'\n", r.Proto))
		}
		if r.DestPort != "" {
			b.WriteString(fmt.Sprintf("\toption dest_port '%s'\n", r.DestPort))
		}
		b.WriteString(fmt.Sprintf("\toption use_policy '%s'\n", r.Policy))
		b.WriteString("\n")
	}
	return b.String()
}

// ==================== DHCP ====================

func (h *NetworkHandler) ListDHCPPools(c *gin.Context) {
	var items []model.DHCPPool
	query := h.DB
	if did := c.Query("device_id"); did != "" {
		query = query.Where("device_id = ?", did)
	}
	query.Find(&items)
	c.JSON(http.StatusOK, items)
}

func (h *NetworkHandler) CreateDHCPPool(c *gin.Context) {
	var item model.DHCPPool
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&item)
	c.JSON(http.StatusCreated, item)
}

func (h *NetworkHandler) DeleteDHCPPool(c *gin.Context) {
	h.DB.Delete(&model.DHCPPool{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *NetworkHandler) ListStaticLeases(c *gin.Context) {
	var items []model.StaticLease
	query := h.DB
	if did := c.Query("device_id"); did != "" {
		query = query.Where("device_id = ?", did)
	}
	query.Find(&items)
	c.JSON(http.StatusOK, items)
}

func (h *NetworkHandler) CreateStaticLease(c *gin.Context) {
	var item model.StaticLease
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&item)
	c.JSON(http.StatusCreated, item)
}

func (h *NetworkHandler) DeleteStaticLease(c *gin.Context) {
	h.DB.Delete(&model.StaticLease{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ==================== VLAN ====================

func (h *NetworkHandler) ListVLANs(c *gin.Context) {
	var items []model.VLAN
	query := h.DB
	if did := c.Query("device_id"); did != "" {
		query = query.Where("device_id = ?", did)
	}
	query.Order("vid").Find(&items)
	c.JSON(http.StatusOK, items)
}

func (h *NetworkHandler) CreateVLAN(c *gin.Context) {
	var item model.VLAN
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&item)
	c.JSON(http.StatusCreated, item)
}

func (h *NetworkHandler) UpdateVLAN(c *gin.Context) {
	var item model.VLAN
	if err := h.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vlan not found"})
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

func (h *NetworkHandler) DeleteVLAN(c *gin.Context) {
	h.DB.Delete(&model.VLAN{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ==================== Apply DHCP ====================

func (h *NetworkHandler) ApplyDHCP(c *gin.Context) {
	deviceID := c.Param("device_id")
	var device model.Device
	if err := h.DB.First(&device, deviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	var pools []model.DHCPPool
	h.DB.Where("device_id = ? AND enabled = true", deviceID).Find(&pools)
	var leases []model.StaticLease
	h.DB.Where("device_id = ?", deviceID).Find(&leases)

	uci := generateDHCPUCI(pools, leases)

	if h.MQTT != nil && h.MQTT.IsConnected() {
		topic := fmt.Sprintf("nexusgate/devices/%s/config", device.MAC)
		h.MQTT.Publish(topic, 1, false, uci)
	}
	h.DB.Create(&model.DeviceConfig{DeviceID: device.ID, Content: uci, Status: "pending"})
	writeAudit(h.DB, c, "apply", "dhcp", fmt.Sprintf("applied DHCP config to device %s", device.Name))
	c.JSON(http.StatusOK, gin.H{"message": "dhcp config pushed", "uci": uci})
}

func generateDHCPUCI(pools []model.DHCPPool, leases []model.StaticLease) string {
	var b strings.Builder
	for _, p := range pools {
		b.WriteString(fmt.Sprintf("config dhcp '%s'\n", p.Interface))
		b.WriteString(fmt.Sprintf("\toption interface '%s'\n", p.Interface))
		b.WriteString(fmt.Sprintf("\toption start '%d'\n", p.Start))
		b.WriteString(fmt.Sprintf("\toption limit '%d'\n", p.Limit))
		b.WriteString(fmt.Sprintf("\toption leasetime '%s'\n", p.LeaseTime))
		if p.DNS != "" {
			for _, dns := range strings.Split(p.DNS, ",") {
				dns = strings.TrimSpace(dns)
				if dns != "" {
					b.WriteString(fmt.Sprintf("\tlist dhcp_option '6,%s'\n", dns))
				}
			}
		}
		if p.Gateway != "" {
			b.WriteString(fmt.Sprintf("\tlist dhcp_option '3,%s'\n", p.Gateway))
		}
		b.WriteString("\n")
	}
	for _, l := range leases {
		b.WriteString("config host\n")
		b.WriteString(fmt.Sprintf("\toption name '%s'\n", l.Name))
		b.WriteString(fmt.Sprintf("\toption mac '%s'\n", l.MAC))
		b.WriteString(fmt.Sprintf("\toption ip '%s'\n", l.IP))
		b.WriteString("\n")
	}
	return b.String()
}

// ==================== Apply VLAN ====================

func (h *NetworkHandler) ApplyVLAN(c *gin.Context) {
	deviceID := c.Param("device_id")
	var device model.Device
	if err := h.DB.First(&device, deviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	var vlans []model.VLAN
	h.DB.Where("device_id = ?", deviceID).Order("vid").Find(&vlans)

	uci := generateVLANUCI(vlans)

	if h.MQTT != nil && h.MQTT.IsConnected() {
		topic := fmt.Sprintf("nexusgate/devices/%s/config", device.MAC)
		h.MQTT.Publish(topic, 1, false, uci)
	}
	h.DB.Create(&model.DeviceConfig{DeviceID: device.ID, Content: uci, Status: "pending"})
	writeAudit(h.DB, c, "apply", "vlan", fmt.Sprintf("applied VLAN config to device %s", device.Name))
	c.JSON(http.StatusOK, gin.H{"message": "vlan config pushed", "uci": uci})
}

func generateVLANUCI(vlans []model.VLAN) string {
	var b strings.Builder
	for _, v := range vlans {
		// Bridge VLAN filtering entry
		b.WriteString(fmt.Sprintf("config bridge-vlan 'brvlan%d'\n", v.VID))
		b.WriteString("\toption device 'br-lan'\n")
		b.WriteString(fmt.Sprintf("\toption vlan '%d'\n", v.VID))
		b.WriteString("\tlist ports 'lan1:t'\n")
		b.WriteString("\tlist ports 'lan2:t'\n")
		b.WriteString("\n")

		// Network interface for the VLAN
		ifName := v.Name
		if ifName == "" {
			ifName = fmt.Sprintf("vlan%d", v.VID)
		}
		b.WriteString(fmt.Sprintf("config interface '%s'\n", ifName))
		b.WriteString("\toption proto 'static'\n")
		device := v.Interface
		if device == "" {
			device = fmt.Sprintf("br-lan.%d", v.VID)
		}
		b.WriteString(fmt.Sprintf("\toption device '%s'\n", device))
		if v.IPAddr != "" {
			b.WriteString(fmt.Sprintf("\toption ipaddr '%s'\n", v.IPAddr))
		}
		b.WriteString(fmt.Sprintf("\toption netmask '%s'\n", v.Netmask))
		b.WriteString("\n")

		// If isolated, add a firewall zone
		if v.Isolated {
			b.WriteString(fmt.Sprintf("# firewall: zone '%s' isolated — deny forwarding\n", ifName))
			b.WriteString(fmt.Sprintf("config zone '%s_zone'\n", ifName))
			b.WriteString(fmt.Sprintf("\toption name '%s'\n", ifName))
			b.WriteString(fmt.Sprintf("\tlist network '%s'\n", ifName))
			b.WriteString("\toption input 'ACCEPT'\n")
			b.WriteString("\toption output 'ACCEPT'\n")
			b.WriteString("\toption forward 'REJECT'\n")
			b.WriteString("\n")
		}
	}
	return b.String()
}
