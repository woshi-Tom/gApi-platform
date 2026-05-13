package handler

import "github.com/gin-gonic/gin"

// extractUserID safely extracts user_id from gin context.
// Returns (userID, true) on success, (0, false) if missing or wrong type.
func extractUserID(c *gin.Context) (uint, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}
