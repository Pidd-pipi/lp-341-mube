package util

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明。
type Claims struct {
	UserID uint   `json:"user_id"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT。
func GenerateToken(secret string, expireHours int, userID uint, phone, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID, Phone: phone, Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "gbcheckup",
			Subject:   strconv.FormatUint(uint64(userID), 10),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken 解析 JWT。
func ParseToken(secret, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
