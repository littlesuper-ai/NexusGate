package main

import (
	"log"

	"github.com/nexusgate/nexusgate/internal/config"
	"github.com/nexusgate/nexusgate/internal/handler"
	"github.com/nexusgate/nexusgate/internal/mqtt"
	"github.com/nexusgate/nexusgate/internal/store"
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

	mqttClient, err := mqtt.NewClient(cfg)
	if err != nil {
		log.Printf("warning: MQTT connection failed: %v", err)
	}

	r := handler.SetupRouter(db, mqttClient, cfg)

	log.Printf("NexusGate server starting on %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
