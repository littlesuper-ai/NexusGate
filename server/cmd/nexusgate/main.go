package main

import (
	"log"

	"github.com/nexusgate/nexusgate/internal/config"
	"github.com/nexusgate/nexusgate/internal/handler"
	"github.com/nexusgate/nexusgate/internal/jobs"
	"github.com/nexusgate/nexusgate/internal/mqtt"
	"github.com/nexusgate/nexusgate/internal/store"
	"github.com/nexusgate/nexusgate/internal/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := store.InitDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := store.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	store.SeedAdminUser(db)

	mqttClient, err := mqtt.NewClient(cfg)
	if err != nil {
		log.Printf("warning: MQTT connection failed: %v", err)
	}

	wsHub := ws.NewHub(cfg.JWTSecret)

	if mqttClient != nil {
		mqtt.SubscribeDeviceStatus(mqttClient, db, wsHub)
		mqtt.SubscribeConfigACK(mqttClient, db, wsHub)
		mqtt.SubscribeUpgradeACK(mqttClient, db, wsHub)
	}

	// Start background jobs
	jobs.StartOfflineDetector(db, wsHub)
	jobs.StartMetricsCleanup(db)

	r := handler.SetupRouter(db, mqttClient, cfg, wsHub)

	log.Printf("NexusGate server starting on %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
