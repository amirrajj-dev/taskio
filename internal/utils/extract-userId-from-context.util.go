package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


func GetUserIDFromContext(c *gin.Context) (*uuid.UUID, bool) {
    res, exists := c.Get("auth.userId")
    if !exists {
        return nil, false
    }

    id, ok := res.(uuid.UUID)
    if !ok {
        return nil, false
    }

    return &id, true
}
