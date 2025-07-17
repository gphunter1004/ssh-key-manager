package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// InstallPublicKeyToServer는 공개키를 로컬 서버의 authorized_keys에 추가합니다.
func InstallPublicKeyToServer(publicKey, sshUser, sshHomePath string) error {
	log.Printf("🔧 로컬 서버에 SSH 공개키 설치 시작")
	log.Printf("   - SSH 사용자: %s", sshUser)
	log.Printf("   - SSH 홈 경로: %s", sshHomePath)

	// SSH 디렉토리 경로 구성
	sshDir := filepath.Join(sshHomePath, ".ssh")
	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")

	// .ssh 디렉토리 생성 (존재하지 않는 경우)
	if err := ensureSSHDirectory(sshDir); err != nil {
		return fmt.Errorf("SSH 디렉토리 생성 실패: %v", err)
	}

	// authorized_keys 파일에 공개키 추가
	if err := appendToAuthorizedKeys(authorizedKeysPath, publicKey); err != nil {
		return fmt.Errorf("authorized_keys 업데이트 실패: %v", err)
	}

	log.Printf("✅ SSH 공개키 설치 완료")
	log.Printf("   - 파일 경로: %s", authorizedKeysPath)
	return nil
}

// RemovePublicKeyFromServer는 공개키를 로컬 서버의 authorized_keys에서 제거합니다.
func RemovePublicKeyFromServer(publicKeyToRemove, sshUser, sshHomePath string) error {
	log.Printf("🗑️ 로컬 서버에서 SSH 공개키 제거 시작")

	authorizedKeysPath := filepath.Join(sshHomePath, ".ssh", "authorized_keys")

	// 파일이 존재하지 않으면 제거할 것이 없음
	if _, err := os.Stat(authorizedKeysPath); os.IsNotExist(err) {
		log.Printf("⚠️ authorized_keys 파일이 존재하지 않습니다: %s", authorizedKeysPath)
		return nil
	}

	// 기존 파일 읽기
	lines, err := readAuthorizedKeysFile(authorizedKeysPath)
	if err != nil {
		return fmt.Errorf("authorized_keys 파일 읽기 실패: %v", err)
	}

	// 제거할 키와 일치하는 라인 찾기 및 제거
	publicKeyToRemove = strings.TrimSpace(publicKeyToRemove)
	var filteredLines []string
	removed := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			// 빈 줄이나 주석은 유지
			filteredLines = append(filteredLines, line)
			continue
		}

		// 공개키 비교 (코멘트 부분 제외하고 키 부분만 비교)
		if !isSamePublicKey(line, publicKeyToRemove) {
			filteredLines = append(filteredLines, line)
		} else {
			removed = true
			log.Printf("   - 제거된 키: %s", truncateString(line, 50))
		}
	}

	if !removed {
		log.Printf("⚠️ 제거할 키를 찾을 수 없습니다")
		return nil
	}

	// 파일 다시 쓰기
	if err := writeAuthorizedKeysFile(authorizedKeysPath, filteredLines); err != nil {
		return fmt.Errorf("authorized_keys 파일 쓰기 실패: %v", err)
	}

	log.Printf("✅ SSH 공개키 제거 완료")
	return nil
}

// ensureSSHDirectory는 .ssh 디렉토리가 존재하는지 확인하고 생성합니다.
func ensureSSHDirectory(sshDir string) error {
	// 디렉토리가 이미 존재하는지 확인
	if info, err := os.Stat(sshDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s는 디렉토리가 아닙니다", sshDir)
		}
		log.Printf("   - .ssh 디렉토리 존재함: %s", sshDir)
		return nil
	}

	// 디렉토리 생성 (권한: 700)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return err
	}

	log.Printf("   - .ssh 디렉토리 생성: %s", sshDir)
	return nil
}

// appendToAuthorizedKeys는 authorized_keys 파일에 공개키를 추가합니다.
func appendToAuthorizedKeys(authorizedKeysPath, publicKey string) error {
	publicKey = strings.TrimSpace(publicKey)

	// 기존 키 중복 체크
	if exists, err := isKeyAlreadyExists(authorizedKeysPath, publicKey); err != nil {
		return err
	} else if exists {
		log.Printf("   - 이미 존재하는 키입니다 (건너뜀)")
		return nil
	}

	// 파일에 추가 (없으면 생성)
	file, err := os.OpenFile(authorizedKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// 파일 끝에 개행이 없으면 추가
	if err := ensureFileEndsWithNewline(authorizedKeysPath); err != nil {
		log.Printf("⚠️ 개행 문자 추가 실패: %v", err)
	}

	// 공개키 추가
	if _, err := file.WriteString(publicKey + "\n"); err != nil {
		return err
	}

	log.Printf("   - 공개키 추가: %s", truncateString(publicKey, 50))
	return nil
}

// isKeyAlreadyExists는 공개키가 이미 authorized_keys에 존재하는지 확인합니다.
func isKeyAlreadyExists(authorizedKeysPath, newPublicKey string) (bool, error) {
	// 파일이 존재하지 않으면 키도 존재하지 않음
	if _, err := os.Stat(authorizedKeysPath); os.IsNotExist(err) {
		return false, nil
	}

	lines, err := readAuthorizedKeysFile(authorizedKeysPath)
	if err != nil {
		return false, err
	}

	newPublicKey = strings.TrimSpace(newPublicKey)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if isSamePublicKey(line, newPublicKey) {
			return true, nil
		}
	}

	return false, nil
}

// isSamePublicKey는 두 공개키가 같은지 비교합니다 (코멘트 제외).
func isSamePublicKey(key1, key2 string) bool {
	// 키를 공백으로 분할하여 알고리즘과 키 부분만 비교
	parts1 := strings.Fields(strings.TrimSpace(key1))
	parts2 := strings.Fields(strings.TrimSpace(key2))

	// 최소한 알고리즘과 키 데이터가 있어야 함
	if len(parts1) < 2 || len(parts2) < 2 {
		return false
	}

	// 알고리즘과 키 데이터가 같은지 확인 (코멘트는 제외)
	return parts1[0] == parts2[0] && parts1[1] == parts2[1]
}

// readAuthorizedKeysFile은 authorized_keys 파일을 읽어서 라인별로 반환합니다.
func readAuthorizedKeysFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// writeAuthorizedKeysFile은 authorized_keys 파일에 라인들을 씁니다.
func writeAuthorizedKeysFile(path string, lines []string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// ensureFileEndsWithNewline은 파일이 개행 문자로 끝나는지 확인하고 필요시 추가합니다.
func ensureFileEndsWithNewline(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// 파일 크기 확인
	info, err := file.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		return nil // 빈 파일이면 개행 불필요
	}

	// 마지막 바이트 확인
	if _, err := file.Seek(-1, 2); err != nil {
		return err
	}

	lastByte := make([]byte, 1)
	if _, err := file.Read(lastByte); err != nil {
		return err
	}

	// 마지막이 개행이 아니면 추가
	if lastByte[0] != '\n' {
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}

	return nil
}

// truncateString은 문자열을 지정된 길이로 자르고 ...을 추가합니다.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ValidateSSHConfig는 SSH 설정이 유효한지 확인합니다.
func ValidateSSHConfig(sshUser, sshHomePath string) error {
	if sshUser == "" {
		return fmt.Errorf("SSH_USER가 설정되지 않았습니다")
	}
	if sshHomePath == "" {
		return fmt.Errorf("SSH_HOME_PATH가 설정되지 않았습니다")
	}

	// 홈 디렉토리 존재 확인
	if info, err := os.Stat(sshHomePath); err != nil {
		return fmt.Errorf("SSH 홈 디렉토리에 접근할 수 없습니다: %s (%v)", sshHomePath, err)
	} else if !info.IsDir() {
		return fmt.Errorf("SSH 홈 경로가 디렉토리가 아닙니다: %s", sshHomePath)
	}

	return nil
}
