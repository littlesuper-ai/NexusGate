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

type VPNHandler struct {
	DB   *gorm.DB
	MQTT mqtt.Client
}

// --- WireGuard Interfaces ---

func (h *VPNHandler) ListInterfaces(c *gin.Context) {
	deviceID := c.Query("device_id")
	var ifaces []model.WireGuardInterface
	query := h.DB
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	query.Find(&ifaces)
	c.JSON(http.StatusOK, ifaces)
}

func (h *VPNHandler) CreateInterface(c *gin.Context) {
	var iface model.WireGuardInterface
	if err := c.ShouldBindJSON(&iface); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&iface).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, iface)
}

func (h *VPNHandler) UpdateInterface(c *gin.Context) {
	var iface model.WireGuardInterface
	if err := h.DB.First(&iface, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "interface not found"})
		return
	}
	if err := c.ShouldBindJSON(&iface); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&iface)
	c.JSON(http.StatusOK, iface)
}

func (h *VPNHandler) DeleteInterface(c *gin.Context) {
	h.DB.Where("interface_id = ?", c.Param("id")).Delete(&model.WireGuardPeer{})
	if err := h.DB.Delete(&model.WireGuardInterface{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- WireGuard Peers ---

func (h *VPNHandler) ListPeers(c *gin.Context) {
	ifaceID := c.Query("interface_id")
	var peers []model.WireGuardPeer
	query := h.DB
	if ifaceID != "" {
		query = query.Where("interface_id = ?", ifaceID)
	}
	query.Find(&peers)
	c.JSON(http.StatusOK, peers)
}

func (h *VPNHandler) CreatePeer(c *gin.Context) {
	var peer model.WireGuardPeer
	if err := c.ShouldBindJSON(&peer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&peer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, peer)
}

func (h *VPNHandler) UpdatePeer(c *gin.Context) {
	var peer model.WireGuardPeer
	if err := h.DB.First(&peer, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "peer not found"})
		return
	}
	if err := c.ShouldBindJSON(&peer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&peer)
	c.JSON(http.StatusOK, peer)
}

func (h *VPNHandler) DeletePeer(c *gin.Context) {
	if err := h.DB.Delete(&model.WireGuardPeer{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ApplyVPN generates WireGuard UCI config and pushes to device.
func (h *VPNHandler) ApplyVPN(c *gin.Context) {
	deviceID := c.Param("device_id")

	var device model.Device
	if err := h.DB.First(&device, deviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	var ifaces []model.WireGuardInterface
	h.DB.Where("device_id = ? AND enabled = true", deviceID).Find(&ifaces)

	var allPeers []model.WireGuardPeer
	for _, iface := range ifaces {
		var peers []model.WireGuardPeer
		h.DB.Where("interface_id = ? AND enabled = true", iface.ID).Find(&peers)
		allPeers = append(allPeers, peers...)
	}

	uci := generateWireGuardUCI(ifaces, allPeers)

	if h.MQTT != nil && h.MQTT.IsConnected() {
		topic := fmt.Sprintf("nexusgate/devices/%s/config", device.MAC)
		h.MQTT.Publish(topic, 1, false, uci)
	}

	h.DB.Create(&model.DeviceConfig{
		DeviceID: device.ID,
		Content:  uci,
		Status:   "pending",
	})

	c.JSON(http.StatusOK, gin.H{"message": "vpn config pushed", "uci": uci})
}

func generateWireGuardUCI(ifaces []model.WireGuardInterface, peers []model.WireGuardPeer) string {
	var b strings.Builder

	for _, iface := range ifaces {
		b.WriteString(fmt.Sprintf("config interface '%s'\n", iface.Name))
		b.WriteString("\toption proto 'wireguard'\n")
		b.WriteString(fmt.Sprintf("\toption private_key '%s'\n", iface.PrivateKey))
		b.WriteString(fmt.Sprintf("\tlist addresses '%s'\n", iface.Address))
		b.WriteString(fmt.Sprintf("\toption listen_port '%d'\n", iface.ListenPort))
		b.WriteString("\n")

		for _, peer := range peers {
			if peer.InterfaceID != iface.ID {
				continue
			}
			b.WriteString(fmt.Sprintf("config wireguard_%s\n", iface.Name))
			if peer.Description != "" {
				b.WriteString(fmt.Sprintf("\toption description '%s'\n", peer.Description))
			}
			b.WriteString(fmt.Sprintf("\toption public_key '%s'\n", peer.PublicKey))
			for _, cidr := range strings.Split(peer.AllowedIPs, ",") {
				cidr = strings.TrimSpace(cidr)
				if cidr != "" {
					b.WriteString(fmt.Sprintf("\tlist allowed_ips '%s'\n", cidr))
				}
			}
			if peer.Endpoint != "" {
				b.WriteString(fmt.Sprintf("\toption endpoint_host '%s'\n", peer.Endpoint))
			}
			if peer.Keepalive > 0 {
				b.WriteString(fmt.Sprintf("\toption persistent_keepalive '%d'\n", peer.Keepalive))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}
