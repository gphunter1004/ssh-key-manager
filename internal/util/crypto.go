package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

var (
	// 글로벌 JWT 시크릿 (한 번만 초기화)
	jwtSecret         string
	secretInitialized bool
)

// ========== JWT 관련 ==========

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
		jwtSecret = generateSecureSecret()
		log.Printf("⚠️ JWT 시크릿이 설정되지 않아 임시 시크릿을 생성했습니다")
	}

	secretInitialized = true
}

// GetJWTSecret은 안전하게 JWT 시크릿을 반환합니다.
func GetJWTSecret() (string, error) {
	if !secretInitialized {
		return "", fmt.Errorf("JWT secret not initialized")
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
		return fmt.Sprintf("fallback-secret-%d-%d", os.Getpid(), time.Now().Unix())
	}
	return fmt.Sprintf("%x", bytes)
}

// GenerateJWT는 JWT 토큰을 생성합니다.
func GenerateJWT(userID uint) (string, error) {
	if userID == 0 {
		return "", fmt.Errorf("userID cannot be zero")
	}

	secret, err := GetJWTSecret()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Minute)),
			Issuer:    "ssh-key-manager",
			Subject:   fmt.Sprintf("user-%d", userID),
			ID:        fmt.Sprintf("%d-%d", userID, now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT는 JWT 토큰을 검증합니다.
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token string is empty")
	}

	secret, err := GetJWTSecret()
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("invalid token")
	}

	if claims.UserID == 0 {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	if claims.Issuer != "ssh-key-manager" {
		return nil, fmt.Errorf("invalid token issuer")
	}

	return claims, nil
}

// ========== 비밀번호 해싱 ==========

// HashPassword는 비밀번호를 해시합니다.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	if len(password) > 72 {
		return "", fmt.Errorf("password too long (max 72 characters)")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash는 비밀번호와 해시를 비교합니다.
func CheckPasswordHash(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ========== SSH 키 생성 통합 API ==========

// SSHKeyPair는 생성된 SSH 키 쌍을 담는 구조체입니다.
type SSHKeyPair struct {
	PrivateKeyPEM []byte // PEM 형식 개인키
	PublicKeySSH  []byte // SSH authorized_keys 형식 공개키
	PPKKey        []byte // PuTTY PPK 형식 개인키
	Algorithm     string // 키 알고리즘 (RSA, Ed25519 등)
	Bits          int    // 키 크기
}

// GenerateSSHKeyPair는 완전한 SSH 키 쌍을 생성합니다.
func GenerateSSHKeyPair(bits int, comment string) (*SSHKeyPair, error) {
	// 기본값 설정
	if bits <= 0 {
		bits = 4096
	}
	if comment == "" {
		comment = "ssh-key-manager"
	}

	// 안전한 comment 처리
	comment = sanitizeComment(comment)

	// RSA 키 생성
	privateKey, err := generateRSAPrivateKey(bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %v", err)
	}

	// PEM 형식 개인키
	pemKey, err := encodePrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PEM: %v", err)
	}

	// SSH 공개키
	publicKey, err := generateSSHPublicKey(privateKey, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SSH public key: %v", err)
	}

	// PPK 형식
	ppkKey, err := generatePPKKey(privateKey, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PPK: %v", err)
	}

	return &SSHKeyPair{
		PrivateKeyPEM: pemKey,
		PublicKeySSH:  publicKey,
		PPKKey:        ppkKey,
		Algorithm:     "RSA",
		Bits:          bits,
	}, nil
}

// ========== 내부 헬퍼 함수들 ==========

// generateRSAPrivateKey는 RSA 개인키를 생성합니다.
func generateRSAPrivateKey(bits int) (*rsa.PrivateKey, error) {
	if bits < 2048 {
		return nil, fmt.Errorf("key size too small: minimum 2048 bits required")
	}
	if bits > 8192 {
		return nil, fmt.Errorf("key size too large: maximum 8192 bits allowed")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	if err := privateKey.Validate(); err != nil {
		return nil, fmt.Errorf("generated key is invalid: %v", err)
	}

	return privateKey, nil
}

// encodePrivateKeyToPEM은 개인키를 PEM 형식으로 인코딩합니다.
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	if err := privateKey.Validate(); err != nil {
		return nil, fmt.Errorf("private key validation failed: %v", err)
	}

	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
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

// generateSSHPublicKey는 SSH 공개키를 생성합니다.
func generateSSHPublicKey(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

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

// generatePPKKey는 PuTTY PPK 형식 키를 생성합니다.
func generatePPKKey(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	// SSH 공개키 생성
	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %v", err)
	}

	// 공개키를 SSH wire format으로 인코딩
	publicKeyBytes := sshPublicKey.Marshal()

	// 개인키를 SSH wire format으로 인코딩
	privateKeyBytes, err := marshalRSAPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %v", err)
	}

	// Base64 인코딩
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKeyBytes)
	privateKeyB64 := base64.StdEncoding.EncodeToString(privateKeyBytes)

	// 70자씩 줄바꿈
	publicLines := splitBase64(publicKeyB64, 70)
	privateLines := splitBase64(privateKeyB64, 70)

	// PPK v2 형식
	ppkContent := fmt.Sprintf(`PuTTY-User-Key-File-2: ssh-rsa
Encryption: none
Comment: %s
Public-Lines: %d
%s
Private-Lines: %d
%s
Private-MAC: %s
`,
		comment,
		len(publicLines),
		strings.Join(publicLines, "\n"),
		len(privateLines),
		strings.Join(privateLines, "\n"),
		generateSimpleMAC(publicKeyBytes, privateKeyBytes),
	)

	return []byte(ppkContent), nil
}

// marshalRSAPrivateKey는 RSA 개인키를 SSH wire format으로 마샬링합니다.
func marshalRSAPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	d := privateKey.D.Bytes()
	p := privateKey.Primes[0].Bytes()
	q := privateKey.Primes[1].Bytes()

	// iqmp = q^-1 mod p
	qInv := new(big.Int).ModInverse(privateKey.Primes[1], privateKey.Primes[0])
	iqmp := qInv.Bytes()

	var result []byte
	result = append(result, marshalMpint(d)...)
	result = append(result, marshalMpint(p)...)
	result = append(result, marshalMpint(q)...)
	result = append(result, marshalMpint(iqmp)...)

	return result, nil
}

// marshalMpint는 SSH wire format으로 정수를 마샬링합니다.
func marshalMpint(data []byte) []byte {
	if len(data) > 0 && data[0]&0x80 != 0 {
		data = append([]byte{0}, data...)
	}

	length := uint32(len(data))
	result := make([]byte, 4+len(data))
	result[0] = byte(length >> 24)
	result[1] = byte(length >> 16)
	result[2] = byte(length >> 8)
	result[3] = byte(length)
	copy(result[4:], data)

	return result
}

// splitBase64는 Base64 문자열을 지정된 길이로 분할합니다.
func splitBase64(data string, lineLength int) []string {
	var lines []string
	for i := 0; i < len(data); i += lineLength {
		end := i + lineLength
		if end > len(data) {
			end = len(data)
		}
		lines = append(lines, data[i:end])
	}
	return lines
}

// generateSimpleMAC는 간단한 MAC을 생성합니다.
func generateSimpleMAC(publicKey, privateKey []byte) string {
	sum := 0
	for _, b := range publicKey {
		sum += int(b)
	}
	for _, b := range privateKey {
		sum += int(b)
	}
	return fmt.Sprintf("%064x", sum)
}

// sanitizeComment는 안전한 comment를 만듭니다.
func sanitizeComment(comment string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == ' ' {
			return r
		}
		return -1
	}, comment)
}
