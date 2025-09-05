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

// ========== JWT 관련 단순화 ==========

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

// GenerateJWT는 JWT 토큰을 생성합니다 (단순화).
func GenerateJWT(userID uint) (string, error) {
	if userID == 0 {
		return "", fmt.Errorf("userID cannot be zero")
	}

	if !secretInitialized || jwtSecret == "" {
		return "", fmt.Errorf("JWT secret not initialized")
	}

	now := time.Now()

	// 단순한 클레임만 사용
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(24 * time.Hour).Unix(), // 24시간
		"iat":     now.Unix(),
		"iss":     "ssh-key-manager",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
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

// ========== SSH 키 생성 ==========

// SSHKeyPair는 생성된 SSH 키 쌍을 담는 구조체입니다.
type SSHKeyPair struct {
	PrivateKeyPEM []byte // PEM 형식 개인키
	PublicKeySSH  []byte // SSH authorized_keys 형식 공개키
	PPKKey        []byte // PuTTY PPK 형식 개인키
}

// GenerateSSHKeyPair는 SSH 키 쌍을 생성합니다 (개선됨).
func GenerateSSHKeyPair(bits int, comment string) (*SSHKeyPair, error) {
	log.Printf("🔐 SSH 키 생성 시작: %d bits, comment: %s", bits, comment)

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
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %v", err)
	}

	// PEM 형식 개인키 생성
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	}
	pemKey := pem.EncodeToMemory(privBlock)

	// PEM 키 검증
	if len(pemKey) == 0 {
		return nil, fmt.Errorf("failed to encode PEM private key")
	}
	log.Printf("✅ PEM 개인키 생성 완료: %d bytes", len(pemKey))

	// SSH 공개키 생성
	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %v", err)
	}

	authorizedKey := ssh.MarshalAuthorizedKey(publicRsaKey)
	keyStr := strings.TrimSuffix(string(authorizedKey), "\n")
	publicKey := []byte(keyStr + " " + comment + "\n")

	// 공개키 검증
	if len(publicKey) == 0 {
		return nil, fmt.Errorf("failed to generate SSH public key")
	}
	log.Printf("✅ SSH 공개키 생성 완료: %d bytes", len(publicKey))

	// PPK 형식 생성
	ppkKey, err := generatePPKKey(privateKey, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PPK: %v", err)
	}

	// PPK 키 검증
	if len(ppkKey) == 0 {
		return nil, fmt.Errorf("failed to generate PPK key")
	}
	log.Printf("✅ PPK 키 생성 완료: %d bytes", len(ppkKey))

	result := &SSHKeyPair{
		PrivateKeyPEM: pemKey,
		PublicKeySSH:  publicKey,
		PPKKey:        ppkKey,
	}

	// 최종 검증
	if err := validateSSHKeyPair(result); err != nil {
		return nil, fmt.Errorf("key pair validation failed: %v", err)
	}

	log.Printf("✅ SSH 키 쌍 생성 완료 - PEM: %d, Public: %d, PPK: %d bytes",
		len(result.PrivateKeyPEM), len(result.PublicKeySSH), len(result.PPKKey))

	return result, nil
}

// validateSSHKeyPair 생성된 키 쌍의 유효성을 검사합니다.
func validateSSHKeyPair(keyPair *SSHKeyPair) error {
	if keyPair == nil {
		return fmt.Errorf("keyPair is nil")
	}

	if len(keyPair.PrivateKeyPEM) == 0 {
		return fmt.Errorf("private key PEM is empty")
	}

	if len(keyPair.PublicKeySSH) == 0 {
		return fmt.Errorf("public key SSH is empty")
	}

	if len(keyPair.PPKKey) == 0 {
		return fmt.Errorf("PPK key is empty")
	}

	// PEM 형식 검증
	if !strings.Contains(string(keyPair.PrivateKeyPEM), "BEGIN RSA PRIVATE KEY") {
		return fmt.Errorf("invalid PEM format")
	}

	// SSH 공개키 형식 검증
	if !strings.HasPrefix(string(keyPair.PublicKeySSH), "ssh-rsa ") {
		return fmt.Errorf("invalid SSH public key format")
	}

	// PPK 형식 검증
	if !strings.Contains(string(keyPair.PPKKey), "PuTTY-User-Key-File") {
		return fmt.Errorf("invalid PPK format")
	}

	return nil
}

// generatePPKKey는 PuTTY PPK 형식 키를 생성합니다.
func generatePPKKey(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
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

	result := []byte(ppkContent)
	if len(result) == 0 {
		return nil, fmt.Errorf("PPK content is empty")
	}

	return result, nil
}

// marshalRSAPrivateKey는 RSA 개인키를 SSH wire format으로 마샬링합니다.
func marshalRSAPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	if len(privateKey.Primes) < 2 {
		return nil, fmt.Errorf("invalid private key: insufficient primes")
	}

	d := privateKey.D.Bytes()
	p := privateKey.Primes[0].Bytes()
	q := privateKey.Primes[1].Bytes()

	// iqmp = q^-1 mod p
	qInv := new(big.Int).ModInverse(privateKey.Primes[1], privateKey.Primes[0])
	if qInv == nil {
		return nil, fmt.Errorf("failed to compute modular inverse")
	}
	iqmp := qInv.Bytes()

	var result []byte
	result = append(result, marshalMpint(d)...)
	result = append(result, marshalMpint(p)...)
	result = append(result, marshalMpint(q)...)
	result = append(result, marshalMpint(iqmp)...)

	if len(result) == 0 {
		return nil, fmt.Errorf("marshaled key is empty")
	}

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
	if comment == "" {
		return "ssh-key-manager"
	}

	// 안전한 문자만 유지
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == ' ' {
			return r
		}
		return -1
	}, comment)

	// 빈 문자열이면 기본값 반환
	if strings.TrimSpace(result) == "" {
		return "ssh-key-manager"
	}

	return result
}

// ========== 내부 헬퍼 함수들 ==========

// generateSecureSecret은 안전한 랜덤 시크릿을 생성합니다.
func generateSecureSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("fallback-secret-%d-%d", os.Getpid(), time.Now().Unix())
	}
	return fmt.Sprintf("%x", bytes)
}
