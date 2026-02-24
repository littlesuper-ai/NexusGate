package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nexusgate/nexusgate/internal/config"
	"github.com/nexusgate/nexusgate/internal/handler/middleware"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, mqttClient mqtt.Client, cfg *config.Config) *gin.Engine {
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

		// Dashboard
		api.GET("/dashboard/summary", deviceHandler.DashboardSummary)
	}

	return r
}
