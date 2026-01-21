package jwt

import (
	"fmt"
	"time"

	"github.com/Rudraksh121a/BookStore/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string, cfg *config.Config) (string, error) {
	var jwtSecret = []byte(cfg.JWT_SECRET)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24h expiry
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func VerifyJWT(tokenStr string, cfg *config.Config) (*jwt.Token, error) {
	var jwtSecret = []byte(cfg.JWT_SECRET)
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
}
