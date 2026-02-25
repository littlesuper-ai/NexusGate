package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
)

// HealthCheck returns a handler that checks DB and MQTT connectivity.
func HealthCheck(db *gorm.DB, mqttClient mqtt.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "ok"
		httpCode := http.StatusOK
		checks := gin.H{}

		// Database check
		sqlDB, err := db.DB()
		if err != nil {
			checks["database"] = "error: " + err.Error()
			status = "degraded"
			httpCode = http.StatusServiceUnavailable
		} else if err := sqlDB.Ping(); err != nil {
			checks["database"] = "error: " + err.Error()
			status = "degraded"
			httpCode = http.StatusServiceUnavailable
		} else {
			checks["database"] = "ok"
		}

		// MQTT check
		if mqttClient == nil {
			checks["mqtt"] = "not configured"
			status = "degraded"
		} else if mqttClient.IsConnected() {
			checks["mqtt"] = "ok"
		} else {
			checks["mqtt"] = "disconnected"
			status = "degraded"
		}

		c.JSON(httpCode, gin.H{
			"status": status,
			"checks": checks,
		})
	}
}
