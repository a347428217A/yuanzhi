package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// CustomerClaims 用户端JWT声明
type CustomerClaims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

// MerchantClaims 商家端JWT声明
var secret = []byte("admin-go-api")

type MerchantClaims struct {
	AdminID    uint   `json:"admin_id"`
	MerchantID uint   `json:"merchant_id"`
	Role       string `json:"role"`
	jwt.StandardClaims
}

// GenerateCustomerToken 生成用户端JWT
func GenerateCustomerToken(userID uint) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &CustomerClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseCustomerToken 解析用户端JWT
func ParseCustomerToken(tokenString string) (*CustomerClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomerClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if claims, ok := token.Claims.(*CustomerClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

// GenerateMerchantToken 生成商家端JWT
func GenerateMerchantToken(adminID, merchantID uint, role, secret string) (string, error) {
	expirationTime := time.Now().Add(8 * time.Hour) // 商家端token有效期8小时

	claims := &MerchantClaims{
		AdminID:    adminID,
		MerchantID: merchantID,
		Role:       role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseMerchantToken 解析商家端JWT
func ParseMerchantToken(tokenString, secret string) (*MerchantClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MerchantClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if claims, ok := token.Claims.(*MerchantClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	// 实际项目中应该使用bcrypt或argon2
	// 这里简化处理
	return password, nil
}

// CheckPassword 检查密码
func CheckPassword(hashedPassword, password string) bool {
	return hashedPassword == password
}
