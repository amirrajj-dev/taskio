package helpers

import (
	"net/http"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/gin-gonic/gin"
)

func ShouldBindJSON(c *gin.Context , dto interface{}) bool {
	if err := c.ShouldBindJSON(dto);err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest , errors.NewBasicError("invalid request payload : " + err.Error() , c.Request.URL.Path))
		return false
	}
	return true
}