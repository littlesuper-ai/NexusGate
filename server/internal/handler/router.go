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

	// Protected routes — all authenticated users (including viewer) can read
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Read-only routes (viewer, operator, admin)
		api.GET("/devices", deviceHandler.List)
		api.GET("/devices/:id", deviceHandler.Get)
		api.GET("/devices/:id/metrics", deviceHandler.Metrics)
		api.GET("/devices/:id/config/history", configHandler.ConfigHistory)
		api.GET("/templates", configHandler.ListTemplates)
		api.GET("/firewall/zones", firewallHandler.ListZones)
		api.GET("/firewall/rules", firewallHandler.ListRules)
		api.GET("/vpn/interfaces", vpnHandler.ListInterfaces)
		api.GET("/vpn/peers", vpnHandler.ListPeers)
		api.GET("/firmware", firmwareHandler.List)
		api.GET("/firmware/download/:filename", firmwareHandler.Download)
		api.GET("/firmware/upgrades", firmwareHandler.UpgradeHistory)
		api.GET("/network/wan", networkHandler.ListWANInterfaces)
		api.GET("/network/mwan/policies", networkHandler.ListMWANPolicies)
		api.GET("/network/mwan/rules", networkHandler.ListMWANRules)
		api.GET("/network/dhcp/pools", networkHandler.ListDHCPPools)
		api.GET("/network/dhcp/leases", networkHandler.ListStaticLeases)
		api.GET("/network/vlans", networkHandler.ListVLANs)
		api.GET("/settings", settingHandler.List)
		api.GET("/settings/:key", settingHandler.Get)
		api.GET("/alerts", alertHandler.List)
		api.GET("/alerts/summary", alertHandler.Summary)
		api.GET("/dashboard/summary", deviceHandler.DashboardSummary)

		// Write routes (operator + admin only)
		write := api.Group("")
		write.Use(middleware.RequireRole("admin", "operator"))
		{
			// Devices
			write.PUT("/devices/:id", deviceHandler.Update)
			write.DELETE("/devices/:id", deviceHandler.Delete)
			write.POST("/devices/:id/reboot", deviceHandler.Reboot)

			// Config
			write.POST("/templates", configHandler.CreateTemplate)
			write.PUT("/templates/:id", configHandler.UpdateTemplate)
			write.DELETE("/templates/:id", configHandler.DeleteTemplate)
			write.POST("/devices/:id/config/push", configHandler.PushConfig)

			// Firewall
			write.POST("/firewall/zones", firewallHandler.CreateZone)
			write.PUT("/firewall/zones/:id", firewallHandler.UpdateZone)
			write.DELETE("/firewall/zones/:id", firewallHandler.DeleteZone)
			write.POST("/firewall/rules", firewallHandler.CreateRule)
			write.PUT("/firewall/rules/:id", firewallHandler.UpdateRule)
			write.DELETE("/firewall/rules/:id", firewallHandler.DeleteRule)
			write.POST("/firewall/apply/:device_id", firewallHandler.ApplyFirewall)

			// VPN
			write.POST("/vpn/interfaces", vpnHandler.CreateInterface)
			write.PUT("/vpn/interfaces/:id", vpnHandler.UpdateInterface)
			write.DELETE("/vpn/interfaces/:id", vpnHandler.DeleteInterface)
			write.POST("/vpn/peers", vpnHandler.CreatePeer)
			write.PUT("/vpn/peers/:id", vpnHandler.UpdatePeer)
			write.DELETE("/vpn/peers/:id", vpnHandler.DeletePeer)
			write.POST("/vpn/apply/:device_id", vpnHandler.ApplyVPN)

			// Firmware
			write.POST("/firmware/upload", firmwareHandler.Upload)
			write.DELETE("/firmware/:id", firmwareHandler.Delete)
			write.POST("/firmware/:id/stable", firmwareHandler.MarkStable)
			write.POST("/firmware/upgrade", firmwareHandler.PushUpgrade)
			write.POST("/firmware/upgrade/batch", firmwareHandler.BatchUpgrade)

			// Multi-WAN
			write.POST("/network/wan", networkHandler.CreateWANInterface)
			write.PUT("/network/wan/:id", networkHandler.UpdateWANInterface)
			write.DELETE("/network/wan/:id", networkHandler.DeleteWANInterface)
			write.POST("/network/mwan/policies", networkHandler.CreateMWANPolicy)
			write.PUT("/network/mwan/policies/:id", networkHandler.UpdateMWANPolicy)
			write.DELETE("/network/mwan/policies/:id", networkHandler.DeleteMWANPolicy)
			write.POST("/network/mwan/rules", networkHandler.CreateMWANRule)
			write.PUT("/network/mwan/rules/:id", networkHandler.UpdateMWANRule)
			write.DELETE("/network/mwan/rules/:id", networkHandler.DeleteMWANRule)
			write.POST("/network/mwan/apply/:device_id", networkHandler.ApplyMWAN)

			// DHCP
			write.POST("/network/dhcp/pools", networkHandler.CreateDHCPPool)
			write.PUT("/network/dhcp/pools/:id", networkHandler.UpdateDHCPPool)
			write.DELETE("/network/dhcp/pools/:id", networkHandler.DeleteDHCPPool)
			write.POST("/network/dhcp/leases", networkHandler.CreateStaticLease)
			write.PUT("/network/dhcp/leases/:id", networkHandler.UpdateStaticLease)
			write.DELETE("/network/dhcp/leases/:id", networkHandler.DeleteStaticLease)
			write.POST("/network/dhcp/apply/:device_id", networkHandler.ApplyDHCP)

			// VLAN
			write.POST("/network/vlans", networkHandler.CreateVLAN)
			write.PUT("/network/vlans/:id", networkHandler.UpdateVLAN)
			write.DELETE("/network/vlans/:id", networkHandler.DeleteVLAN)
			write.POST("/network/vlans/apply/:device_id", networkHandler.ApplyVLAN)

			// Settings
			write.POST("/settings", settingHandler.Upsert)
			write.POST("/settings/batch", settingHandler.BatchUpsert)
			write.DELETE("/settings/:key", settingHandler.Delete)

			// Alerts
			write.POST("/alerts/:id/resolve", alertHandler.Resolve)
		}

		// Admin-only routes
		admin := api.Group("")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", authHandler.ListUsers)
			admin.POST("/users", authHandler.CreateUser)
			admin.PUT("/users/:id", authHandler.UpdateUser)
			admin.DELETE("/users/:id", authHandler.DeleteUser)
			admin.GET("/audit-logs", authHandler.AuditLogs)
		}
	}

	return r
}
