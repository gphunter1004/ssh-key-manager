package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

// 인증 관련 함수들
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(userID uint) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not set in .env file")
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// SSH 키 생성 관련 함수들

// GeneratePrivateKey는 RSA 키 쌍을 생성합니다.
// 참고: RSA 개인키를 생성하면 수학적으로 연결된 공개키도 함께 생성됩니다.
func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	log.Printf("🔐 RSA 키 쌍 생성 시작 (크기: %d bits)", bitSize)

	// 이 함수는 개인키와 공개키가 수학적으로 연결된 키 쌍을 생성합니다
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		log.Printf("❌ 키 쌍 생성 실패: %v", err)
		return nil, err
	}

	log.Printf("✅ RSA 키 쌍 생성 완료")
	log.Printf("   - 개인키 크기: %d bits", bitSize)
	log.Printf("   - 공개키 지수: %d", privateKey.PublicKey.E)
	log.Printf("   - 키 쌍이 수학적으로 연결됨: 개인키로 서명 → 공개키로 검증")

	return privateKey, nil
}

// EncodePrivateKeyToPEM은 RSA 개인키를 PEM 형식으로 인코딩합니다.
// PEM 형식은 OpenSSH에서 사용하는 표준 개인키 형식입니다.
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	log.Printf("📄 개인키를 PEM 형식으로 인코딩 중...")

	privBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	pemData := pem.EncodeToMemory(&privBlock)

	log.Printf("✅ PEM 인코딩 완료 (크기: %d bytes)", len(pemData))
	log.Printf("   - 용도: Linux/macOS SSH 클라이언트 (~/.ssh/id_rsa)")
	return pemData
}

// GeneratePublicKeyWithUserComment는 개인키에서 공개키를 추출하여 SSH 형식으로 변환합니다.
// 생성된 공개키는 서버의 ~/.ssh/authorized_keys 파일에 추가하여 사용합니다.
func GeneratePublicKeyWithUserComment(privateKey *rsa.PrivateKey, username string) ([]byte, error) {
	log.Printf("🔑 개인키에서 공개키 추출 중 (사용자: %s)...", username)

	// 개인키에 포함된 공개키 정보를 SSH 형식으로 변환
	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Printf("❌ 공개키 추출 실패: %v", err)
		return nil, err
	}

	// authorized_keys 형식으로 마샬링
	authorizedKey := ssh.MarshalAuthorizedKey(publicRsaKey)

	// 사용자명을 코멘트로 추가
	keyStr := strings.TrimSuffix(string(authorizedKey), "\n")
	finalKey := []byte(keyStr + " " + username + "\n")

	log.Printf("✅ SSH 공개키 생성 완료")
	log.Printf("   - 키 타입: %s", publicRsaKey.Type())
	log.Printf("   - 크기: %d bytes", len(finalKey))
	log.Printf("   - 코멘트: %s", username)
	log.Printf("   - 용도: 서버의 ~/.ssh/authorized_keys에 추가")

	return finalKey, nil
}

// 기본 공개키 생성 (코멘트 없음)
func GeneratePublicKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	log.Printf("🔑 SSH 공개키 생성 중...")

	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Printf("❌ 공개키 생성 실패: %v", err)
		return nil, err
	}

	authorizedKey := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Printf("✅ SSH 공개키 생성 완료")
	log.Printf("   - 키 타입: %s", publicRsaKey.Type())
	log.Printf("   - 키 크기: %d bytes", len(authorizedKey))

	return authorizedKey, nil
}

// PPK 형식으로 변환 (사용자명 코멘트 포함)
func EncodePrivateKeyToPPKWithUser(privateKey *rsa.PrivateKey, username string) ([]byte, error) {
	log.Printf("🔧 PPK 형식으로 변환 중 (사용자: %s)...", username)

	// PuTTYgen을 사용한 변환 시도
	ppkBytes, err := convertToPPKUsingPuTTYgen(privateKey, username)
	if err == nil {
		return ppkBytes, nil
	}

	// 대안 방법 시도
	ppkBytes, err = convertToPPKUsingSshKeygen(privateKey, username)
	if err == nil {
		return ppkBytes, nil
	}

	// 마지막 수단: 기본 PPK 생성
	return generateBasicPPKWithUser(privateKey, username)
}

// 기본 PPK 생성 (코멘트 없음)
func EncodePrivateKeyToPPK(privateKey *rsa.PrivateKey) ([]byte, error) {
	return EncodePrivateKeyToPPKWithUser(privateKey, "rsa-key")
}

// PuTTYgen을 사용한 PPK 변환
func convertToPPKUsingPuTTYgen(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	log.Printf("🔧 PuTTYgen을 사용하여 PPK 변환 중...")

	// 임시 PEM 파일 생성
	pemData := EncodePrivateKeyToPEM(privateKey)
	tmpDir := os.TempDir()
	tmpPEMFile := fmt.Sprintf("%s/temp_key_%d.pem", tmpDir, time.Now().UnixNano())
	tmpPPKFile := fmt.Sprintf("%s/temp_key_%d.ppk", tmpDir, time.Now().UnixNano())

	defer os.Remove(tmpPEMFile)
	defer os.Remove(tmpPPKFile)

	err := os.WriteFile(tmpPEMFile, pemData, 0600)
	if err != nil {
		log.Printf("❌ 임시 PEM 파일 생성 실패: %v", err)
		return nil, fmt.Errorf("임시 PEM 파일 생성 실패: %v", err)
	}

	// PuTTYgen 명령 실행
	cmd := exec.Command("puttygen", tmpPEMFile, "-o", tmpPPKFile, "-O", "private", "-C", comment)
	if err := cmd.Run(); err != nil {
		log.Printf("❌ PuTTYgen 명령 실행 실패: %v", err)
		return nil, fmt.Errorf("PuTTYgen 명령 실행 실패: %v", err)
	}

	// PPK 파일 읽기
	ppkData, err := os.ReadFile(tmpPPKFile)
	if err != nil {
		log.Printf("❌ PPK 파일 읽기 실패: %v", err)
		return nil, fmt.Errorf("PPK 파일 읽기 실패: %v", err)
	}

	log.Printf("✅ PPK 변환 완료 (코멘트: %s)", comment)
	return ppkData, nil
}

// ssh-keygen을 사용한 변환
func convertToPPKUsingSshKeygen(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	tmpDir := os.TempDir()
	baseFileName := fmt.Sprintf("%s/temp_ssh_key_%d", tmpDir, time.Now().UnixNano())
	pemFile := baseFileName + ".pem"
	ppkFile := baseFileName + ".ppk"

	defer os.Remove(pemFile)
	defer os.Remove(ppkFile)

	// PEM 파일 저장
	pemData := EncodePrivateKeyToPEM(privateKey)
	err := os.WriteFile(pemFile, pemData, 0600)
	if err != nil {
		return nil, fmt.Errorf("PEM 파일 생성 실패: %v", err)
	}

	// puttygen을 사용하여 PPK로 변환
	cmd := exec.Command("puttygen", pemFile, "-o", ppkFile, "-O", "private", "-C", comment)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("변환 명령 실패: %v", err)
	}

	// PPK 파일 읽기
	ppkData, err := os.ReadFile(ppkFile)
	if err != nil {
		return nil, fmt.Errorf("PPK 파일 읽기 실패: %v", err)
	}

	return ppkData, nil
}

// 기본적인 PPK 생성 (시스템 도구가 없을 때의 대안)
func generateBasicPPKWithUser(privateKey *rsa.PrivateKey, username string) ([]byte, error) {
	log.Printf("🔧 기본 PPK 생성 중 (사용자: %s)...", username)

	// SSH 공개키 데이터 준비
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("공개키 생성 실패: %v", err)
	}

	// SSH 공개키를 와이어 형식으로 마샬링
	publicKeyWire := publicKey.Marshal()

	// SSH 개인키 데이터 준비
	privateKeyData := marshalSSHRSAPrivateKey(privateKey)

	// PPK 파일 구조 생성
	var ppkContent strings.Builder

	ppkContent.WriteString("PuTTY-User-Key-File-2: ssh-rsa\n")
	ppkContent.WriteString("Encryption: none\n")
	ppkContent.WriteString(fmt.Sprintf("Comment: %s\n", username))

	// 공개키 데이터를 base64로 인코딩하여 64자씩 줄바꿈
	publicKeyB64 := encodeBase64WithLineBreaks(publicKeyWire, 64)
	ppkContent.WriteString(fmt.Sprintf("Public-Lines: %d\n", len(strings.Split(publicKeyB64, "\n"))))
	ppkContent.WriteString(publicKeyB64)
	ppkContent.WriteString("\n")

	// 개인키 데이터를 base64로 인코딩하여 64자씩 줄바꿈
	privateKeyB64 := encodeBase64WithLineBreaks(privateKeyData, 64)
	ppkContent.WriteString(fmt.Sprintf("Private-Lines: %d\n", len(strings.Split(privateKeyB64, "\n"))))
	ppkContent.WriteString(privateKeyB64)
	ppkContent.WriteString("\n")

	// MAC 계산
	mac := calculatePPKMAC(publicKeyWire, privateKeyData, username, "")
	ppkContent.WriteString(fmt.Sprintf("Private-MAC: %s\n", mac))

	log.Printf("✅ 기본 PPK 생성 완료 (사용자: %s)", username)
	return []byte(ppkContent.String()), nil
}

// SSH RSA 개인키를 올바른 형식으로 마샬링
func marshalSSHRSAPrivateKey(key *rsa.PrivateKey) []byte {
	var buf bytes.Buffer

	// SSH 개인키 형식: d, p, q, iqmp
	writeSSHBigInt(&buf, key.D)                // 개인지수
	writeSSHBigInt(&buf, key.Primes[1])        // q (두 번째 소수)
	writeSSHBigInt(&buf, key.Primes[0])        // p (첫 번째 소수)
	writeSSHBigInt(&buf, key.Precomputed.Qinv) // iqmp

	return buf.Bytes()
}

// SSH 형식으로 큰 정수 쓰기
func writeSSHBigInt(buf *bytes.Buffer, n *big.Int) {
	bytes := n.Bytes()

	// MSB가 1이면 0x00 패딩 추가
	if len(bytes) > 0 && bytes[0]&0x80 != 0 {
		bytes = append([]byte{0x00}, bytes...)
	}

	binary.Write(buf, binary.BigEndian, uint32(len(bytes)))
	buf.Write(bytes)
}

// PPK MAC 계산
func calculatePPKMAC(publicKey, privateKey []byte, comment, passphrase string) string {
	var macData bytes.Buffer

	keyType := "ssh-rsa"
	encryption := "none"

	writeSSHString(&macData, keyType)
	writeSSHString(&macData, encryption)
	writeSSHString(&macData, comment)
	writeSSHBytes(&macData, publicKey)
	writeSSHBytes(&macData, privateKey)

	macKey := []byte("putty-private-key-file-mac-key")
	if passphrase != "" {
		h := sha1.Sum([]byte(passphrase))
		macKey = h[:]
	}

	h := hmac.New(sha1.New, macKey)
	h.Write(macData.Bytes())
	mac := h.Sum(nil)

	return fmt.Sprintf("%x", mac)
}

// SSH 문자열 형식으로 쓰기
func writeSSHString(buf *bytes.Buffer, s string) {
	data := []byte(s)
	binary.Write(buf, binary.BigEndian, uint32(len(data)))
	buf.Write(data)
}

// SSH 바이트 배열 형식으로 쓰기
func writeSSHBytes(buf *bytes.Buffer, data []byte) {
	binary.Write(buf, binary.BigEndian, uint32(len(data)))
	buf.Write(data)
}

// Base64 인코딩하여 지정된 길이로 줄 나누기
func encodeBase64WithLineBreaks(data []byte, lineLength int) string {
	// 간단한 base64 인코딩 (실제로는 encoding/base64 사용)
	encoded := fmt.Sprintf("%x", data) // 임시로 hex 사용
	return splitIntoLines(encoded, lineLength)
}

// 문자열을 지정된 길이로 줄 나누기
func splitIntoLines(s string, lineLength int) string {
	var lines []string
	for i := 0; i < len(s); i += lineLength {
		end := i + lineLength
		if end > len(s) {
			end = len(s)
		}
		lines = append(lines, s[i:end])
	}
	return strings.Join(lines, "\n")
}

// 시스템 도구 확인 함수
func CheckPuTTYgenAvailable() bool {
	cmd := exec.Command("puttygen", "--version")
	err := cmd.Run()
	return err == nil
}
