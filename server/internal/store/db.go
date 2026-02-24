package store

import (
	"github.com/nexusgate/nexusgate/internal/config"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Device{},
		&model.DeviceMetrics{},
		&model.ConfigTemplate{},
		&model.DeviceConfig{},
		&model.AuditLog{},
		&model.FirewallZone{},
		&model.FirewallRule{},
		&model.WireGuardInterface{},
		&model.WireGuardPeer{},
		&model.Firmware{},
		&model.FirmwareUpgrade{},
		&model.WANInterface{},
		&model.MWANPolicy{},
		&model.MWANRule{},
		&model.DHCPPool{},
		&model.StaticLease{},
		&model.VLAN{},
		&model.SystemSetting{},
		&model.Alert{},
	)
}
