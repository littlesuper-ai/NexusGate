package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
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

	wsHub := ws.NewHub(cfg.JWTSecret)

	// Subscribe handler — called on initial connect and on every reconnect
	subscribeAll := func(c pahomqtt.Client) {
		mqtt.SubscribeDeviceStatus(c, db, wsHub)
		mqtt.SubscribeConfigACK(c, db, wsHub)
		mqtt.SubscribeUpgradeACK(c, db, wsHub)
	}

	mqttClient, err := mqtt.NewClient(cfg, subscribeAll)
	if err != nil {
		log.Printf("warning: MQTT connection failed: %v", err)
	}

	// Start background jobs
	jobs.StartOfflineDetector(db, wsHub)
	jobs.StartMetricsCleanup(db)
	jobs.StartAutoUpgradeChecker(db, mqttClient)

	r := handler.SetupRouter(db, mqttClient, cfg, wsHub)

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("NexusGate server starting on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	// Close MQTT connection
	if mqttClient != nil && mqttClient.IsConnected() {
		mqttClient.Disconnect(1000)
	}

	// Close database connection
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	log.Println("server exited")
}
