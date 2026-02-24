package handler

import (
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

	db.Create(&model.AuditLog{
		UserID:   uid,
		Username: uname,
		Action:   action,
		Resource: resource,
		Detail:   detail,
		IP:       c.ClientIP(),
	})
}
