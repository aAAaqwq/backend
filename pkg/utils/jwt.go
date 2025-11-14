package utils

import (
	"backend/config"
	"backend/internal/model"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-change-in-production") // TODO: 从配置文件读取

// JWTClaims JWT声明
type JWTClaims struct {
	UID  int64  `json:"uid"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func LoadJWTSecret(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWT.Secret)
}

// GenerateToken 生成JWT token
func GenerateToken(user *model.User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // token有效期24小时

	claims := JWTClaims{
		UID:  user.UID,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "sensor_manage_hub",
			Subject:   ConvertToString(user.UID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// SetJWTSecret 设置JWT密钥（从配置读取）
func SetJWTSecret(secret string) {
	if secret != "" {
		jwtSecret = []byte(secret)
	}
}
