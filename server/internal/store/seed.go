package store

import (
	"log"

	"github.com/nexusgate/nexusgate/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdminUser(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("warning: failed to hash seed password: %v", err)
		return
	}

	admin := model.User{
		Username: "admin",
		Password: string(hashed),
		Role:     model.RoleAdmin,
		Email:    "admin@nexusgate.local",
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("warning: failed to seed admin user: %v", err)
		return
	}

	log.Println("seeded default admin user (admin / admin123)")
}
