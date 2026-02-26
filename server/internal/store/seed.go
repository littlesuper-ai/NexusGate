package store

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

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

	// Use ADMIN_PASSWORD env var or generate a random secure password
	password := os.Getenv("ADMIN_PASSWORD")
	generated := false
	if password == "" {
		b := make([]byte, 16)
		rand.Read(b)
		password = hex.EncodeToString(b)
		generated = true
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

	if generated {
		log.Printf("seeded default admin user — username: admin, password: %s", password)
		log.Println("IMPORTANT: change the admin password immediately or set ADMIN_PASSWORD env var")
	} else {
		log.Println("seeded admin user with password from ADMIN_PASSWORD env var")
	}
}
