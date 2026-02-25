package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/config"
	"github.com/nexusgate/nexusgate/internal/handler/middleware"
	"github.com/nexusgate/nexusgate/internal/ws"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, mqttClient mqtt.Client, cfg *config.Config, wsHub *ws.Hub) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	authHandler := &AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	deviceHandler := &DeviceHandler{DB: db, MQTT: mqttClient}
	configHandler := &ConfigHandler{DB: db, MQTT: mqttClient}
	firewallHandler := &FirewallHandler{DB: db, MQTT: mqttClient}
	vpnHandler := &VPNHandler{DB: db, MQTT: mqttClient}
	firmwareHandler := &FirmwareHandler{DB: db, MQTT: mqttClient}
	networkHandler := &NetworkHandler{DB: db, MQTT: mqttClient}
	settingHandler := &SettingHandler{DB: db}
	alertHandler := &AlertHandler{DB: db}

	// WebSocket endpoint (no JWT for WS upgrade, auth via query param)
	r.GET("/ws", wsHub.HandleWS)

	// Public routes
	pub := r.Group("/api/v1")
	{
		pub.POST("/auth/login", authHandler.Login)
		pub.POST("/devices/register", deviceHandler.Register)
	}

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Devices
		api.GET("/devices", deviceHandler.List)
		api.GET("/devices/:id", deviceHandler.Get)
		api.PUT("/devices/:id", deviceHandler.Update)
		api.DELETE("/devices/:id", deviceHandler.Delete)
		api.POST("/devices/:id/reboot", deviceHandler.Reboot)
		api.GET("/devices/:id/metrics", deviceHandler.Metrics)

		// Config templates
		api.GET("/templates", configHandler.ListTemplates)
		api.POST("/templates", configHandler.CreateTemplate)
		api.PUT("/templates/:id", configHandler.UpdateTemplate)
		api.DELETE("/templates/:id", configHandler.DeleteTemplate)

		// Config deployment
		api.POST("/devices/:id/config/push", configHandler.PushConfig)
		api.GET("/devices/:id/config/history", configHandler.ConfigHistory)

		// Users (admin only)
		admin := api.Group("")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", authHandler.ListUsers)
			admin.POST("/users", authHandler.CreateUser)
			admin.DELETE("/users/:id", authHandler.DeleteUser)
			admin.GET("/audit-logs", authHandler.AuditLogs)
		}

		// Firewall
		api.GET("/firewall/zones", firewallHandler.ListZones)
		api.POST("/firewall/zones", firewallHandler.CreateZone)
		api.PUT("/firewall/zones/:id", firewallHandler.UpdateZone)
		api.DELETE("/firewall/zones/:id", firewallHandler.DeleteZone)
		api.GET("/firewall/rules", firewallHandler.ListRules)
		api.POST("/firewall/rules", firewallHandler.CreateRule)
		api.PUT("/firewall/rules/:id", firewallHandler.UpdateRule)
		api.DELETE("/firewall/rules/:id", firewallHandler.DeleteRule)
		api.POST("/firewall/apply/:device_id", firewallHandler.ApplyFirewall)

		// VPN (WireGuard)
		api.GET("/vpn/interfaces", vpnHandler.ListInterfaces)
		api.POST("/vpn/interfaces", vpnHandler.CreateInterface)
		api.PUT("/vpn/interfaces/:id", vpnHandler.UpdateInterface)
		api.DELETE("/vpn/interfaces/:id", vpnHandler.DeleteInterface)
		api.GET("/vpn/peers", vpnHandler.ListPeers)
		api.POST("/vpn/peers", vpnHandler.CreatePeer)
		api.PUT("/vpn/peers/:id", vpnHandler.UpdatePeer)
		api.DELETE("/vpn/peers/:id", vpnHandler.DeletePeer)
		api.POST("/vpn/apply/:device_id", vpnHandler.ApplyVPN)

		// Firmware
		api.GET("/firmware", firmwareHandler.List)
		api.POST("/firmware/upload", firmwareHandler.Upload)
		api.GET("/firmware/download/:filename", firmwareHandler.Download)
		api.DELETE("/firmware/:id", firmwareHandler.Delete)
		api.POST("/firmware/:id/stable", firmwareHandler.MarkStable)
		api.POST("/firmware/upgrade", firmwareHandler.PushUpgrade)
		api.POST("/firmware/upgrade/batch", firmwareHandler.BatchUpgrade)
		api.GET("/firmware/upgrades", firmwareHandler.UpgradeHistory)

		// Multi-WAN
		api.GET("/network/wan", networkHandler.ListWANInterfaces)
		api.POST("/network/wan", networkHandler.CreateWANInterface)
		api.PUT("/network/wan/:id", networkHandler.UpdateWANInterface)
		api.DELETE("/network/wan/:id", networkHandler.DeleteWANInterface)
		api.GET("/network/mwan/policies", networkHandler.ListMWANPolicies)
		api.POST("/network/mwan/policies", networkHandler.CreateMWANPolicy)
		api.PUT("/network/mwan/policies/:id", networkHandler.UpdateMWANPolicy)
		api.DELETE("/network/mwan/policies/:id", networkHandler.DeleteMWANPolicy)
		api.GET("/network/mwan/rules", networkHandler.ListMWANRules)
		api.POST("/network/mwan/rules", networkHandler.CreateMWANRule)
		api.PUT("/network/mwan/rules/:id", networkHandler.UpdateMWANRule)
		api.DELETE("/network/mwan/rules/:id", networkHandler.DeleteMWANRule)
		api.POST("/network/mwan/apply/:device_id", networkHandler.ApplyMWAN)

		// DHCP
		api.GET("/network/dhcp/pools", networkHandler.ListDHCPPools)
		api.POST("/network/dhcp/pools", networkHandler.CreateDHCPPool)
		api.PUT("/network/dhcp/pools/:id", networkHandler.UpdateDHCPPool)
		api.DELETE("/network/dhcp/pools/:id", networkHandler.DeleteDHCPPool)
		api.GET("/network/dhcp/leases", networkHandler.ListStaticLeases)
		api.POST("/network/dhcp/leases", networkHandler.CreateStaticLease)
		api.PUT("/network/dhcp/leases/:id", networkHandler.UpdateStaticLease)
		api.DELETE("/network/dhcp/leases/:id", networkHandler.DeleteStaticLease)
		api.POST("/network/dhcp/apply/:device_id", networkHandler.ApplyDHCP)

		// VLAN
		api.GET("/network/vlans", networkHandler.ListVLANs)
		api.POST("/network/vlans", networkHandler.CreateVLAN)
		api.PUT("/network/vlans/:id", networkHandler.UpdateVLAN)
		api.DELETE("/network/vlans/:id", networkHandler.DeleteVLAN)
		api.POST("/network/vlans/apply/:device_id", networkHandler.ApplyVLAN)

		// System settings
		api.GET("/settings", settingHandler.List)
		api.GET("/settings/:key", settingHandler.Get)
		api.POST("/settings", settingHandler.Upsert)
		api.POST("/settings/batch", settingHandler.BatchUpsert)
		api.DELETE("/settings/:key", settingHandler.Delete)

		// Alerts
		api.GET("/alerts", alertHandler.List)
		api.GET("/alerts/summary", alertHandler.Summary)
		api.POST("/alerts/:id/resolve", alertHandler.Resolve)

		// Dashboard
		api.GET("/dashboard/summary", deviceHandler.DashboardSummary)
	}

	return r
}
