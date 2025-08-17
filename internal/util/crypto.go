package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

// JWTClaims는 JWT 클레임 구조체입니다.
type JWTClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// HashPassword는 비밀번호를 해시합니다.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash는 비밀번호와 해시를 비교합니다.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT는 사용자 ID로 JWT 토큰을 생성합니다.
func GenerateJWT(userID uint) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET 환경변수가 설정되지 않았습니다")
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ssh-key-manager",
			Subject:   fmt.Sprintf("user-%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT는 JWT 토큰을 검증합니다.
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET 환경변수가 설정되지 않았습니다")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("유효하지 않은 토큰")
	}

	return claims, nil
}

// GeneratePrivateKey는 RSA 키 쌍을 생성합니다.
func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bitSize)
}

// EncodePrivateKeyToPEM은 RSA 개인키를 PEM 형식으로 인코딩합니다.
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	privBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	return pem.EncodeToMemory(&privBlock)
}

// GeneratePublicKeyWithComment는 개인키에서 공개키를 추출하여 SSH 형식으로 변환합니다.
func GeneratePublicKeyWithComment(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	authorizedKey := ssh.MarshalAuthorizedKey(publicRsaKey)
	keyStr := strings.TrimSuffix(string(authorizedKey), "\n")
	finalKey := []byte(keyStr + " " + comment + "\n")

	return finalKey, nil
}

// GenerateSimplePPK는 간단한 PPK 형식을 생성합니다.
func GenerateSimplePPK(privateKey *rsa.PrivateKey, comment string) []byte {
	pemData := EncodePrivateKeyToPEM(privateKey)
	
	ppkContent := fmt.Sprintf(`PuTTY-User-Key-File-2: ssh-rsa
Encryption: none
Comment: %s
Public-Lines: 6
Private-Lines: 14
Private-MAC: simplified

# Simplified PPK format for %s
%s`, comment, comment, string(pemData))

	return []byte(ppkContent)
}
