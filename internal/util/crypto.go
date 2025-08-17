package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

var (
	// 글로벌 JWT 시크릿 (한 번만 초기화)
	jwtSecret string
	secretInitialized bool
)

// JWTClaims는 JWT 클레임 구조체입니다.
type JWTClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// InitializeJWTSecret은 JWT 시크릿을 초기화합니다.
func InitializeJWTSecret(configSecret string) {
	if secretInitialized {
		return
	}

	if configSecret != "" {
		jwtSecret = configSecret
		log.Printf("✅ JWT 시크릿이 설정에서 로드되었습니다")
	} else if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		jwtSecret = envSecret
		log.Printf("✅ JWT 시크릿이 환경변수에서 로드되었습니다")
	} else {
		// Fallback: 안전한 랜덤 시크릿 생성
		jwtSecret = generateSecureSecret()
		log.Printf("⚠️ JWT 시크릿이 설정되지 않아 임시 시크릿을 생성했습니다. 운영환경에서는 JWT_SECRET을 설정하세요")
	}
	
	secretInitialized = true
}

// GetJWTSecret은 안전하게 JWT 시크릿을 반환합니다.
func GetJWTSecret() (string, error) {
	if !secretInitialized {
		return "", fmt.Errorf("JWT secret not initialized. Call InitializeJWTSecret() first")
	}
	
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT secret is empty")
	}
	
	return jwtSecret, nil
}

// generateSecureSecret은 안전한 랜덤 시크릿을 생성합니다.
func generateSecureSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("❌ 랜덤 시크릿 생성 실패, PID 기반 fallback 사용: %v", err)
		return fmt.Sprintf("fallback-secret-%d-%d", os.Getpid(), time.Now().Unix())
	}
	
	// Base64 URL 인코딩 (JWT에 안전)
	return fmt.Sprintf("%x", bytes)
}

// HashPassword는 비밀번호를 안전하게 해시합니다.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	
	if len(password) > 72 {
		return "", fmt.Errorf("password too long (max 72 characters)")
	}
	
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	
	return string(bytes), nil
}

// CheckPasswordHash는 비밀번호와 해시를 안전하게 비교합니다.
func CheckPasswordHash(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}
	
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT는 사용자 ID로 안전한 JWT 토큰을 생성합니다.
func GenerateJWT(userID uint) (string, error) {
	if userID == 0 {
		return "", fmt.Errorf("userID cannot be zero")
	}
	
	secret, err := GetJWTSecret()
	if err != nil {
		return "", fmt.Errorf("failed to get JWT secret: %v", err)
	}

	now := time.Now()
	expirationTime := now.Add(24 * time.Hour)

	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Minute)), // 5분 전부터 유효
			Issuer:    "ssh-key-manager",
			Subject:   fmt.Sprintf("user-%d", userID),
			ID:        fmt.Sprintf("%d-%d", userID, now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return tokenString, nil
}

// ValidateJWT는 JWT 토큰을 안전하게 검증합니다.
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token string is empty")
	}
	
	secret, err := GetJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT secret: %v", err)
	}

	// 토큰 파싱 및 검증
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 서명 방법 확인
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	// 클레임 추출 및 검증
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims type")
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	// 추가 검증
	if claims.UserID == 0 {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	if claims.Issuer != "ssh-key-manager" {
		return nil, fmt.Errorf("invalid token issuer: %s", claims.Issuer)
	}

	return claims, nil
}

// GeneratePrivateKey는 안전한 RSA 키 쌍을 생성합니다.
func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// 최소 키 크기 확인
	if bitSize < 2048 {
		return nil, fmt.Errorf("key size too small: minimum 2048 bits required")
	}
	
	if bitSize > 8192 {
		return nil, fmt.Errorf("key size too large: maximum 8192 bits allowed")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %v", err)
	}

	// 키 유효성 검증
	if err := privateKey.Validate(); err != nil {
		return nil, fmt.Errorf("generated key is invalid: %v", err)
	}

	return privateKey, nil
}

// EncodePrivateKeyToPEM은 RSA 개인키를 안전하게 PEM 형식으로 인코딩합니다.
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	// 키 유효성 재확인
	if err := privateKey.Validate(); err != nil {
		return nil, fmt.Errorf("private key validation failed: %v", err)
	}

	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if privKeyBytes == nil {
		return nil, fmt.Errorf("failed to marshal private key")
	}

	privBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	}

	pemBytes := pem.EncodeToMemory(&privBlock)
	if pemBytes == nil {
		return nil, fmt.Errorf("failed to encode private key to PEM")
	}

	return pemBytes, nil
}

// GeneratePublicKeyWithComment는 안전하게 공개키를 생성합니다.
func GeneratePublicKeyWithComment(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	if comment == "" {
		comment = "ssh-key-manager"
	}

	// 안전한 comment 처리 (특수문자 제거)
	comment = strings.Map(func(r rune) rune {
		if r > 127 || r < 32 {
			return -1 // 제거
		}
		return r
	}, comment)

	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %v", err)
	}

	authorizedKey := ssh.MarshalAuthorizedKey(publicRsaKey)
	if len(authorizedKey) == 0 {
		return nil, fmt.Errorf("failed to marshal public key")
	}

	keyStr := strings.TrimSuffix(string(authorizedKey), "\n")
	finalKey := []byte(keyStr + " " + comment + "\n")

	return finalKey, nil
}

// GenerateSimplePPK는 간단한 PPK 형식을 안전하게 생성합니다.
func GenerateSimplePPK(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	pemData, err := EncodePrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key to PEM: %v", err)
	}

	if comment == "" {
		comment = "ssh-key-manager"
	}

	// 안전한 comment 처리
	comment = strings.Map(func(r rune) rune {
		if r > 127 || r < 32 {
			return -1
		}
		return r
	}, comment)
	
	ppkContent := fmt.Sprintf(`PuTTY-User-Key-File-2: ssh-rsa
Encryption: none
Comment: %s
Public-Lines: 6
Private-Lines: 14
Private-MAC: simplified

# Simplified PPK format for %s
%s`, comment, comment, string(pemData))

	return []byte(ppkContent), nil
}
