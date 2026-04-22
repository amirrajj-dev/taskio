package utils

import (
	"net/http"
	"time"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/gin-gonic/gin"
)

func ClearAuthCookie(c *gin.Context) {
	var isSecure bool
	if configs.Configs.App.GO_ENV == "production" {
		isSecure = true
	} else {
		isSecure = false
	}

	cookie := &http.Cookie{
		Name:     configs.Configs.COOKIE_NAME, 
		Value:    "",                         
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,                         
	}
	c.SetCookie(cookie.Name, cookie.Value, int(cookie.MaxAge), cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
}