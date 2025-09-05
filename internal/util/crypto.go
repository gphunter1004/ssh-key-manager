package util

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret         string
	secretInitialized bool
)

// ========== JWT 관련 단순화 ==========

// InitializeJWTSecret은 JWT 시크릿을 초기화합니다.
func InitializeJWTSecret(configSecret string) {
	if secretInitialized {
		return
	}
	if os.Getenv("JWT_SECRET") != "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	} else {
		jwtSecret = configSecret
	}
	if jwtSecret == "" {
		jwtSecret = "temporary-secret-for-dev"
		log.Printf("⚠️ JWT 시크릿이 설정되지 않아 임시 시크릿을 사용합니다.")
	}
	secretInitialized = true
}

func GenerateJWT(userID uint) (string, error) {
	if !secretInitialized {
		return "", fmt.Errorf("JWT secret not initialized")
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ========== SSH 키 생성 ==========
type SSHKeyPair struct {
	PrivateKeyPEM []byte
	PublicKeySSH  []byte
	PPKKey        []byte
}

func GenerateSSHKeyPair(bits int, comment string) (*SSHKeyPair, error) {
	log.Printf("🔐 SSH 키 생성 시작 (puttygen 사용)")
	if bits == 0 {
		bits = 4096
	}

	// 1. RSA 개인키 생성
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("RSA 키 생성 실패: %w", err)
	}

	// 2. PEM 형식으로 인코딩
	pemKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// 3. puttygen을 사용하여 PEM으로부터 PPK와 공개키 생성
	ppkKey, publicKey, err := convertPEMToPPKAndPublicKey(pemKey, comment)
	if err != nil {
		return nil, fmt.Errorf("puttygen 변환 실패: %w", err)
	}

	log.Printf("✅ SSH 키 쌍 생성 완료")
	return &SSHKeyPair{
		PrivateKeyPEM: pemKey,
		PublicKeySSH:  publicKey,
		PPKKey:        ppkKey,
	}, nil
}

// convertPEMToPPKAndPublicKey는 puttygen을 호출하여 PPK와 공개키를 모두 생성합니다.
func convertPEMToPPKAndPublicKey(pemKey []byte, comment string) (ppkKey []byte, publicKey []byte, err error) {
	// 임시 PEM 파일 생성
	tmpFile, err := os.CreateTemp("", "private-key-*.pem")
	if err != nil {
		return nil, nil, fmt.Errorf("임시 PEM 파일 생성 실패: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(pemKey); err != nil {
		return nil, nil, fmt.Errorf("임시 PEM 파일 쓰기 실패: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, nil, fmt.Errorf("임시 PEM 파일 닫기 실패: %w", err)
	}

	// PPK 파일 생성
	ppkOutputFile := tmpFile.Name() + ".ppk"
	defer os.Remove(ppkOutputFile)

	// 1. PPK 생성
	cmdConvert := exec.Command("puttygen", tmpFile.Name(), "-o", ppkOutputFile, "-C", comment)
	var stderrConvert bytes.Buffer
	cmdConvert.Stderr = &stderrConvert
	if err := cmdConvert.Run(); err != nil {
		return nil, nil, fmt.Errorf("PPK 변환 명령어 실행 실패: %s, 에러: %w", stderrConvert.String(), err)
	}

	ppkKey, err = os.ReadFile(ppkOutputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("변환된 PPK 파일 읽기 실패: %w", err)
	}

	// 2. 생성된 PPK로부터 공개키 추출
	cmdExportPub := exec.Command("puttygen", "-L", ppkOutputFile)
	var stderrExport bytes.Buffer
	cmdExportPub.Stderr = &stderrExport

	publicKey, err = cmdExportPub.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("공개키 추출 명령어 실행 실패: %s, 에러: %w", stderrExport.String(), err)
	}

	return ppkKey, bytes.TrimSpace(publicKey), nil
}

// ========== 기존 내부 헬퍼 함수들 (참조용으로 남겨둠) ==========
func sanitizeComment(comment string) string {
	if comment == "" {
		return "ssh-key-manager"
	}
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == ' ' {
			return r
		}
		return -1
	}, comment)
	if strings.TrimSpace(result) == "" {
		return "ssh-key-manager"
	}
	return result
}

func generateSecureSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("fallback-secret-%d-%d", os.Getpid(), time.Now().Unix())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
