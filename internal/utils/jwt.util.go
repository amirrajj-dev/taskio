package utils

import (
	"time"

	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(user *models.UserResponse , jwtExpiry time.Duration , secretKey string) (string , error){
	claims := jwt.MapClaims{
		"id" : user.ID,
		"email" : user.Email,
		"exp" : time.Now().Add(jwtExpiry).Unix(),
		"iat" : time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256 , claims)
	return token.SignedString([]byte(secretKey))
}