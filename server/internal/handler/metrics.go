package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/nexusgate/nexusgate/internal/model"
	"github.com/nexusgate/nexusgate/internal/ws"
	"gorm.io/gorm"
)

type metricsCollector struct {
	db  *gorm.DB
	hub *ws.Hub

	devicesTotal   *prometheus.Desc
	devicesOnline  *prometheus.Desc
	devicesOffline *prometheus.Desc
	alertsOpen     *prometheus.Desc
	alertsWarning  *prometheus.Desc
	alertsCritical *prometheus.Desc
	wsClients      *prometheus.Desc
}

func newMetricsCollector(db *gorm.DB, hub *ws.Hub) *metricsCollector {
	return &metricsCollector{
		db:             db,
		hub:            hub,
		devicesTotal:   prometheus.NewDesc("nexusgate_devices_total", "Total number of devices", nil, nil),
		devicesOnline:  prometheus.NewDesc("nexusgate_devices_online", "Number of online devices", nil, nil),
		devicesOffline: prometheus.NewDesc("nexusgate_devices_offline", "Number of offline devices", nil, nil),
		alertsOpen:     prometheus.NewDesc("nexusgate_alerts_unresolved", "Number of unresolved alerts", nil, nil),
		alertsWarning:  prometheus.NewDesc("nexusgate_alerts_warning", "Number of unresolved warning alerts", nil, nil),
		alertsCritical: prometheus.NewDesc("nexusgate_alerts_critical", "Number of unresolved critical alerts", nil, nil),
		wsClients:      prometheus.NewDesc("nexusgate_websocket_clients", "Number of connected WebSocket clients", nil, nil),
	}
}

func (c *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.devicesTotal
	ch <- c.devicesOnline
	ch <- c.devicesOffline
	ch <- c.alertsOpen
	ch <- c.alertsWarning
	ch <- c.alertsCritical
	ch <- c.wsClients
}

func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	var total, online, offline int64
	c.db.Model(&model.Device{}).Count(&total)
	c.db.Model(&model.Device{}).Where("status = ?", model.StatusOnline).Count(&online)
	c.db.Model(&model.Device{}).Where("status = ?", model.StatusOffline).Count(&offline)

	ch <- prometheus.MustNewConstMetric(c.devicesTotal, prometheus.GaugeValue, float64(total))
	ch <- prometheus.MustNewConstMetric(c.devicesOnline, prometheus.GaugeValue, float64(online))
	ch <- prometheus.MustNewConstMetric(c.devicesOffline, prometheus.GaugeValue, float64(offline))

	var alertsOpen, alertsWarn, alertsCrit int64
	c.db.Model(&model.Alert{}).Where("resolved = false").Count(&alertsOpen)
	c.db.Model(&model.Alert{}).Where("resolved = false AND severity = ?", model.SeverityWarning).Count(&alertsWarn)
	c.db.Model(&model.Alert{}).Where("resolved = false AND severity = ?", model.SeverityCritical).Count(&alertsCrit)

	ch <- prometheus.MustNewConstMetric(c.alertsOpen, prometheus.GaugeValue, float64(alertsOpen))
	ch <- prometheus.MustNewConstMetric(c.alertsWarning, prometheus.GaugeValue, float64(alertsWarn))
	ch <- prometheus.MustNewConstMetric(c.alertsCritical, prometheus.GaugeValue, float64(alertsCrit))

	ch <- prometheus.MustNewConstMetric(c.wsClients, prometheus.GaugeValue, float64(c.hub.ClientCount()))
}

// RegisterMetrics creates a Prometheus registry with NexusGate collectors
// and returns a gin.HandlerFunc that serves the /metrics endpoint.
func RegisterMetrics(db *gorm.DB, hub *ws.Hub) gin.HandlerFunc {
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	reg.MustRegister(newMetricsCollector(db, hub))

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
