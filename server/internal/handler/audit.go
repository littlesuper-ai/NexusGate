package handler

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nexusgate/nexusgate/internal/model"
	"gorm.io/gorm"
)

// writeAudit creates an audit log entry from the current request context.
func writeAudit(db *gorm.DB, c *gin.Context, action, resource, detail string) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	uid, _ := userID.(uint)
	uname, _ := username.(string)

	if err := db.Create(&model.AuditLog{
		UserID:   uid,
		Username: uname,
		Action:   action,
		Resource: resource,
		Detail:   detail,
		IP:       c.ClientIP(),
	}).Error; err != nil {
		log.Printf("failed to write audit log [%s %s]: %v", action, resource, err)
	}
}

// writeLoginAudit creates an audit log for login events (before JWT context is set).
func writeLoginAudit(db *gorm.DB, c *gin.Context, userID uint, username, detail string) {
	if err := db.Create(&model.AuditLog{
		UserID:   userID,
		Username: username,
		Action:   "login",
		Resource: "auth",
		Detail:   detail,
		IP:       c.ClientIP(),
	}).Error; err != nil {
		log.Printf("failed to write login audit log: %v", err)
	}
}
