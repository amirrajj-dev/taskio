package middlewares

import (
	"net/http"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const ctxUserIDKey = "auth.userId"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		taskioToken, err := c.Cookie(configs.Configs.COOKIE_NAME)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewBasicError("unauthorized", c.Request.URL.Path))
			return
		}
		token, tokenErr := jwt.Parse(taskioToken, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(configs.Configs.JWT.JWT_SECRET), nil
		})
		if tokenErr != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewBasicError("invalid token", c.Request.URL.Path))
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewBasicError("invalid token claims", c.Request.URL.Path))
			return
		}

		rawID, ok := claims["id"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewBasicError("invalid user id in token", c.Request.URL.Path))
			return
		}

		userID, err := uuid.Parse(rawID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewBasicError("malformed user id", c.Request.URL.Path))
			return
		}
		user , _ := services.UserService.GetUserByID(c.Request.Context() , userID)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusNotFound , errors.NewBasicError("user not found" , c.Request.URL.Path))
		}
		c.Set(ctxUserIDKey, userID)
		c.Next()
	}
}
