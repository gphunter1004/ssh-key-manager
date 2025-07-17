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

// 기존 함수들
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

func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	log.Printf("🔐 RSA 개인키 생성 시작 (크기: %d bits)", bitSize)

	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		log.Printf("❌ 개인키 생성 실패: %v", err)
		return nil, err
	}

	log.Printf("✅ RSA 개인키 생성 완료 (크기: %d bits)", bitSize)
	log.Printf("   - 모듈러스 크기: %d bytes", privateKey.N.BitLen()/8)
	log.Printf("   - 공개 지수: %d", privateKey.E)

	return privateKey, nil
}

func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	log.Printf("📄 PEM 형식으로 개인키 인코딩 중...")

	privBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	pemData := pem.EncodeToMemory(&privBlock)

	log.Printf("✅ PEM 인코딩 완료 (크기: %d bytes)", len(pemData))
	log.Printf("   - PEM 헤더: %s", privBlock.Type)
	log.Printf("📋 PEM 내용:")
	log.Printf("%s", string(pemData))

	return pemData
}

func GeneratePublicKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	return GeneratePublicKeyWithComment(privateKey, "")
}

// 완전한 공개키 생성 (기본 코멘트 포함)
func GeneratePublicKeyWithDefaultComment(privateKey *rsa.PrivateKey) ([]byte, error) {
	comment := GenerateDefaultComment()
	return GeneratePublicKeyWithComment(privateKey, comment)
}

// 로그인 아이디로 공개키 생성
func GeneratePublicKeyWithLoginID(privateKey *rsa.PrivateKey, loginID string) ([]byte, error) {
	comment := GenerateCommentWithLoginID(loginID)
	return GeneratePublicKeyWithComment(privateKey, comment)
}

func GeneratePublicKeyWithComment(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	log.Printf("🔑 SSH 공개키 생성 중...")

	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Printf("❌ 공개키 생성 실패: %v", err)
		return nil, err
	}

	// 기본 공개키 데이터 생성
	authorizedKey := ssh.MarshalAuthorizedKey(publicRsaKey)

	// 코멘트가 있으면 추가
	if comment != "" {
		// 마지막 개행 문자 제거 후 코멘트 추가
		keyStr := strings.TrimSuffix(string(authorizedKey), "\n")
		finalKey := []byte(keyStr + " " + comment + "\n")

		log.Printf("✅ SSH 공개키 생성 완료 (코멘트: %s)", comment)
		log.Printf("   - 키 타입: %s", publicRsaKey.Type())
		log.Printf("   - 키 크기: %d bytes", len(finalKey))
		log.Printf("📋 공개키 전체 내용:")
		log.Printf("%s", strings.TrimSpace(string(finalKey)))

		return finalKey, nil
	}

	log.Printf("✅ SSH 공개키 생성 완료 (코멘트 없음)")
	log.Printf("   - 키 타입: %s", publicRsaKey.Type())
	log.Printf("   - 키 크기: %d bytes", len(authorizedKey))
	log.Printf("📋 공개키 전체 내용:")
	log.Printf("%s", strings.TrimSpace(string(authorizedKey)))

	return authorizedKey, nil
}

// 사용자 정보를 기반으로 기본 코멘트 생성 (로그인 아이디 사용)
func GenerateDefaultComment() string {
	// 로그인 아이디 가져오기 (다양한 환경 변수 시도)
	loginID := getLoginID()

	comment := loginID
	log.Printf("📝 기본 코멘트 생성: %s", comment)

	return comment
}

// 로그인 아이디를 가져오는 함수
func getLoginID() string {
	// 1. USER 환경변수 (Linux/macOS)
	if user := os.Getenv("USER"); user != "" {
		return user
	}

	// 2. USERNAME 환경변수 (Windows)
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}

	// 3. LOGNAME 환경변수 (일부 Unix 시스템)
	if user := os.Getenv("LOGNAME"); user != "" {
		return user
	}

	// 4. USERPROFILE에서 추출 (Windows)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		parts := strings.Split(userProfile, string(os.PathSeparator))
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// 5. HOME에서 추출 (Linux/macOS)
	if home := os.Getenv("HOME"); home != "" {
		parts := strings.Split(home, string(os.PathSeparator))
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// 6. 기본값
	return "user"
}

// 커스텀 로그인 아이디로 코멘트 생성
func GenerateCommentWithLoginID(loginID string) string {
	if loginID == "" {
		return GenerateDefaultComment()
	}

	log.Printf("📝 커스텀 로그인 아이디로 코멘트 생성: %s", loginID)
	return loginID
}

// 로그인 아이디를 사용한 키 쌍 생성
func GenerateAndSaveKeyPairWithLoginID(bitSize int, baseFilename, loginID string) error {
	comment := GenerateCommentWithLoginID(loginID)
	return GenerateAndSaveKeyPairWithComment(bitSize, baseFilename, comment)
}

// PPK 생성 함수 (시스템 명령어 사용 - 가장 확실한 방법)
func EncodePrivateKeyToPPK(privateKey *rsa.PrivateKey) ([]byte, error) {
	return EncodePrivateKeyToPPKWithComment(privateKey, "rsa-key")
}

func EncodePrivateKeyToPPKWithComment(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	// 방법 1: PuTTYgen 명령어 사용 (가장 확실함)
	ppkBytes, err := convertToPPKUsingPuTTYgen(privateKey, comment)
	if err == nil {
		return ppkBytes, nil
	}

	// 방법 2: ssh-keygen + puttygen 사용
	ppkBytes, err = convertToPPKUsingSshKeygen(privateKey, comment)
	if err == nil {
		return ppkBytes, nil
	}

	// 방법 3: 직접 구현 (최후의 수단)
	return generateBasicPPK(privateKey, comment)
}

// PuTTYgen을 사용한 PPK 생성
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

	log.Printf("   - 임시 PEM 파일: %s", tmpPEMFile)
	log.Printf("   - 대상 PPK 파일: %s", tmpPPKFile)

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

	log.Printf("✅ PPK 변환 완료")
	log.Printf("   - PPK 크기: %d bytes", len(ppkData))
	log.Printf("   - 코멘트: %s", comment)
	log.Printf("📋 PPK 전체 내용:")
	log.Printf("%s", string(ppkData))

	return ppkData, nil
}

// ssh-keygen을 사용한 변환
func convertToPPKUsingSshKeygen(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	// 임시 디렉토리 및 파일 준비
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

	// ssh-keygen을 사용하여 OpenSSH 형식으로 변환 후 puttygen 사용
	cmd := exec.Command("puttygen", pemFile, "-o", ppkFile, "-O", "private")
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
func generateBasicPPK(privateKey *rsa.PrivateKey, comment string) ([]byte, error) {
	// 이 방법은 MAC 검증 문제가 있을 수 있으므로 권장하지 않음
	return nil, errors.New("시스템에 puttygen이 설치되지 않았습니다. puttygen을 설치하거나 다른 방법을 사용하세요")
}

// 시스템 도구 확인 함수
func CheckPuTTYgenAvailable() bool {
	cmd := exec.Command("puttygen", "--version")
	err := cmd.Run()
	return err == nil
}

// PPK 생성 방법 확인 및 정보 제공
func GetPPKGenerationInfo() string {
	var methods []string

	if CheckPuTTYgenAvailable() {
		methods = append(methods, "✓ PuTTYgen 사용 가능")
	} else {
		methods = append(methods, "✗ PuTTYgen 설치 필요")
	}

	info := "PPK 생성 방법:\n"
	info += strings.Join(methods, "\n")

	if !CheckPuTTYgenAvailable() {
		info += "\n\n설치 방법:"
		info += "\n- Ubuntu/Debian: sudo apt-get install putty-tools"
		info += "\n- CentOS/RHEL: sudo yum install putty"
		info += "\n- macOS: brew install putty"
		info += "\n- Windows: https://www.putty.org/ 에서 다운로드"
	}

	return info
}

// SSH RSA 개인키를 올바른 형식으로 마샬링
func marshalSSHRSAPrivateKey(key *rsa.PrivateKey) []byte {
	var buf bytes.Buffer

	// SSH 개인키 형식: d, p, q, iqmp (역순으로 q, p)
	writeSSHBigInt(&buf, key.D)                // 개인지수
	writeSSHBigInt(&buf, key.Primes[1])        // q (두 번째 소수)
	writeSSHBigInt(&buf, key.Primes[0])        // p (첫 번째 소수)
	writeSSHBigInt(&buf, key.Precomputed.Qinv) // iqmp

	return buf.Bytes()
}

// SSH 형식으로 큰 정수 쓰기
func writeSSHBigInt(buf *bytes.Buffer, n *big.Int) {
	bytes := n.Bytes()

	// MSB가 1이면 0x00 패딩 추가 (두의 보수 표현을 위해)
	if len(bytes) > 0 && bytes[0]&0x80 != 0 {
		bytes = append([]byte{0x00}, bytes...)
	}

	// 길이 (4바이트) + 데이터
	binary.Write(buf, binary.BigEndian, uint32(len(bytes)))
	buf.Write(bytes)
}

// PPK MAC 계산 (PuTTY의 알고리즘 구현)
func calculatePPKMAC(publicKey, privateKey []byte, comment, passphrase string) string {
	var macData bytes.Buffer

	// MAC 계산용 데이터 준비
	keyType := "ssh-rsa"
	encryption := "none"

	// 데이터 구성: key-type, encryption, comment, public-key-data, private-plaintext-data
	writeSSHString(&macData, keyType)
	writeSSHString(&macData, encryption)
	writeSSHString(&macData, comment)
	writeSSHBytes(&macData, publicKey)
	writeSSHBytes(&macData, privateKey)

	// HMAC-SHA1으로 MAC 계산 (키는 "putty-private-key-file-mac-key")
	macKey := []byte("putty-private-key-file-mac-key")
	if passphrase != "" {
		// 패스워드가 있으면 SHA1(passphrase) 사용
		h := sha1.Sum([]byte(passphrase))
		macKey = h[:]
	}

	h := hmac.New(sha1.New, macKey)
	h.Write(macData.Bytes())
	mac := h.Sum(nil)

	// 16진수 문자열로 변환
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

// 문자열을 지정된 길이로 줄 나누기
func splitIntoLines(s string, lineLength int) []string {
	var lines []string
	for i := 0; i < len(s); i += lineLength {
		end := i + lineLength
		if end > len(s) {
			end = len(s)
		}
		lines = append(lines, s[i:end])
	}
	return lines
}

// PPK 파일 저장
func SavePPKToFile(privateKey *rsa.PrivateKey, filename string) error {
	return SavePPKToFileWithComment(privateKey, filename, GenerateDefaultComment())
}

// 코멘트와 함께 PPK 파일 저장
func SavePPKToFileWithComment(privateKey *rsa.PrivateKey, filename, comment string) error {
	log.Printf("💾 PPK 파일 저장 중: %s", filename)

	ppkBytes, err := EncodePrivateKeyToPPKWithComment(privateKey, comment)
	if err != nil {
		log.Printf("❌ PPK 파일 저장 실패: %v", err)
		return err
	}

	err = os.WriteFile(filename, ppkBytes, 0600)
	if err != nil {
		log.Printf("❌ 파일 쓰기 실패: %v", err)
		return err
	}

	log.Printf("✅ PPK 파일 저장 완료")
	log.Printf("   - 파일명: %s", filename)
	log.Printf("   - 파일 크기: %d bytes", len(ppkBytes))
	log.Printf("   - 파일 권한: 0600 (owner read/write only)")

	return nil
}

// 키 쌍 생성 및 모든 형식으로 저장 (코멘트 포함)
func GenerateAndSaveKeyPair(bitSize int, baseFilename string) error {
	return GenerateAndSaveKeyPairWithComment(bitSize, baseFilename, "")
}

func GenerateAndSaveKeyPairWithComment(bitSize int, baseFilename, comment string) error {
	log.Printf("🚀 키 쌍 생성 시작")
	log.Printf("   - 키 크기: %d bits", bitSize)
	log.Printf("   - 기본 파일명: %s", baseFilename)

	// 1. 개인키 생성
	privateKey, err := GeneratePrivateKey(bitSize)
	if err != nil {
		return err
	}

	// 기본 코멘트 설정
	if comment == "" {
		comment = GenerateDefaultComment()
		log.Printf("   - 자동 생성된 코멘트: %s", comment)
	} else {
		log.Printf("   - 사용자 지정 코멘트: %s", comment)
	}

	// 2. PEM 형식으로 저장
	log.Printf("📄 PEM 파일 저장 중...")
	pemData := EncodePrivateKeyToPEM(privateKey)
	pemFile := baseFilename + ".pem"
	err = os.WriteFile(pemFile, pemData, 0600)
	if err != nil {
		log.Printf("❌ PEM 파일 저장 실패: %v", err)
		return err
	}
	log.Printf("✅ PEM 파일 저장 완료: %s (%d bytes)", pemFile, len(pemData))

	// 3. PPK 형식으로 저장
	ppkFile := baseFilename + ".ppk"
	err = SavePPKToFileWithComment(privateKey, ppkFile, comment)
	if err != nil {
		return err
	}

	// 4. 공개키 저장 (코멘트 포함)
	log.Printf("🔑 공개키 파일 저장 중...")
	publicKeyData, err := GeneratePublicKeyWithComment(privateKey, comment)
	if err != nil {
		log.Printf("❌ 공개키 생성 실패: %v", err)
		return err
	}

	pubFile := baseFilename + ".pub"
	err = os.WriteFile(pubFile, publicKeyData, 0644)
	if err != nil {
		log.Printf("❌ 공개키 파일 저장 실패: %v", err)
		return err
	}
	log.Printf("✅ 공개키 파일 저장 완료: %s (%d bytes)", pubFile, len(publicKeyData))

	// 5. 생성 완료 요약
	log.Printf("🎉 키 쌍 생성 완료!")
	log.Printf("📁 생성된 파일들:")
	log.Printf("   - %s (RSA 개인키, PEM 형식)", pemFile)
	log.Printf("   - %s (PuTTY 개인키, PPK 형식)", ppkFile)
	log.Printf("   - %s (SSH 공개키, authorized_keys 형식)", pubFile)
	log.Printf("🔐 키 정보:")
	log.Printf("   - 알고리즘: RSA")
	log.Printf("   - 키 크기: %d bits", bitSize)
	log.Printf("   - 코멘트: %s", comment)

	// 6. 생성된 모든 파일 내용 표시
	log.Printf("📋 ========== 생성된 파일 내용 ==========")

	// PEM 파일 내용
	log.Printf("📄 %s 내용:", pemFile)
	log.Printf("%s", string(pemData))

	// PPK 파일 내용
	ppkData, err := os.ReadFile(ppkFile)
	if err == nil {
		log.Printf("🔧 %s 내용:", ppkFile)
		log.Printf("%s", string(ppkData))
	}

	// 공개키 파일 내용
	log.Printf("🔑 %s 내용:", pubFile)
	log.Printf("%s", strings.TrimSpace(string(publicKeyData)))

	log.Printf("📋 ========================================")

	return nil
}

// PPK 파일 검증 함수
func ValidatePPKFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// 기본적인 PPK 형식 검증
	if len(lines) < 5 {
		return errors.New("유효하지 않은 PPK 형식")
	}

	if !strings.HasPrefix(lines[0], "PuTTY-User-Key-File-2:") {
		return errors.New("PPK 파일 헤더가 올바르지 않습니다")
	}

	// MAC 라인 찾기
	var macLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "Private-MAC:") {
			macLine = line
			break
		}
	}

	if macLine == "" {
		return errors.New("MAC이 없습니다")
	}

	return nil
}
