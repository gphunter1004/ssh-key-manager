package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"ssh-key-manager/types"
	"strings"
	"time"
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

// DeploySSHKeyToRemoteServer는 SSH 키를 원격 서버에 배포합니다.
func DeploySSHKeyToRemoteServer(publicKey, host string, port int, username string) error {
	log.Printf("📡 원격 서버 SSH 키 배포 시작")
	log.Printf("   - 대상 서버: %s@%s:%d", username, host, port)

	// 공개키 검증
	if strings.TrimSpace(publicKey) == "" {
		return fmt.Errorf("공개키가 비어있습니다")
	}

	// SSH 연결 테스트
	if err := testSSHConnection(host, port, username); err != nil {
		return fmt.Errorf("SSH 연결 테스트 실패: %v", err)
	}

	// 공개키 배포
	if err := deployPublicKeyViaSSH(publicKey, host, port, username); err != nil {
		return fmt.Errorf("공개키 배포 실패: %v", err)
	}

	log.Printf("✅ SSH 키 배포 완료: %s@%s:%d", username, host, port)
	return nil
}

// testSSHConnection은 SSH 연결을 테스트합니다.
func testSSHConnection(host string, port int, username string) error {
	log.Printf("🔍 SSH 연결 테스트 중...")

	// SSH 연결 테스트 명령
	// ssh -o BatchMode=yes -o ConnectTimeout=10 -p PORT USER@HOST "echo 'connection test'"
	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes", // 비밀번호 프롬프트 비활성화
		"-o", "ConnectTimeout=10", // 연결 타임아웃 10초
		"-o", "StrictHostKeyChecking=no", // 호스트 키 확인 비활성화 (첫 연결시)
		"-o", "UserKnownHostsFile=/dev/null", // known_hosts 파일 사용 안함
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", username, host),
		"echo 'SSH connection test successful'",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ SSH 연결 실패: %s", string(output))
		return fmt.Errorf("SSH 연결 실패 (%s@%s:%d): %v", username, host, port, err)
	}

	log.Printf("✅ SSH 연결 테스트 성공")
	return nil
}

// deployPublicKeyViaSSH는 SSH를 통해 공개키를 원격 서버에 배포합니다.
func deployPublicKeyViaSSH(publicKey, host string, port int, username string) error {
	log.Printf("🔑 공개키 배포 중...")

	// 공개키를 정리 (개행 제거 등)
	cleanedKey := strings.TrimSpace(publicKey)
	if !strings.HasSuffix(cleanedKey, "\n") {
		cleanedKey += "\n"
	}

	// SSH를 통해 authorized_keys에 공개키 추가
	// ssh-copy-id와 유사한 기능을 구현
	sshCommand := fmt.Sprintf(
		`mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo 'Key deployed successfully'`,
		strings.ReplaceAll(cleanedKey, "'", "'\"'\"'"), // 작은따옴표 이스케이프
	)

	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=30",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", username, host),
		sshCommand,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ 공개키 배포 실패: %s", string(output))
		return fmt.Errorf("공개키 배포 실패: %v", err)
	}

	// 배포 성공 확인
	if strings.Contains(string(output), "Key deployed successfully") {
		log.Printf("✅ 공개키 배포 성공")
		return nil
	}

	return fmt.Errorf("공개키 배포 확인 실패: %s", string(output))
}

// TestRemoteServerConnection은 원격 서버 연결을 테스트합니다.
func TestRemoteServerConnection(host string, port int, username string) error {
	log.Printf("🔍 원격 서버 연결 테스트: %s@%s:%d", username, host, port)

	// 연결 테스트 (타임아웃 5초)
	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", username, host),
		"echo 'Connection test successful'",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("연결 실패: %v (출력: %s)", err, string(output))
	}

	if strings.Contains(string(output), "Connection test successful") {
		log.Printf("✅ 연결 테스트 성공")
		return nil
	}

	return fmt.Errorf("연결 테스트 응답 확인 실패: %s", string(output))
}

// RemoveSSHKeyFromRemoteServer는 원격 서버에서 SSH 키를 제거합니다.
func RemoveSSHKeyFromRemoteServer(publicKey, host string, port int, username string) error {
	log.Printf("🗑️ 원격 서버에서 SSH 키 제거 시작")
	log.Printf("   - 대상 서버: %s@%s:%d", username, host, port)

	// 제거할 키 정리
	cleanedKey := strings.TrimSpace(publicKey)
	keyParts := strings.Fields(cleanedKey)
	if len(keyParts) < 2 {
		return fmt.Errorf("유효하지 않은 공개키 형식")
	}

	// 키의 주요 부분 (알고리즘 + 키 데이터)만 사용하여 제거
	keyToRemove := keyParts[0] + " " + keyParts[1]

	// SSH를 통해 authorized_keys에서 키 제거
	sshCommand := fmt.Sprintf(
		`grep -v '%s' ~/.ssh/authorized_keys > ~/.ssh/authorized_keys.tmp && mv ~/.ssh/authorized_keys.tmp ~/.ssh/authorized_keys && echo 'Key removed successfully'`,
		strings.ReplaceAll(keyToRemove, "'", "'\"'\"'"),
	)

	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=30",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", username, host),
		sshCommand,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ SSH 키 제거 실패: %s", string(output))
		return fmt.Errorf("SSH 키 제거 실패: %v", err)
	}

	if strings.Contains(string(output), "Key removed successfully") {
		log.Printf("✅ SSH 키 제거 성공")
		return nil
	}

	return fmt.Errorf("SSH 키 제거 확인 실패: %s", string(output))
}

// GetRemoteServerInfo는 원격 서버의 기본 정보를 조회합니다.
func GetRemoteServerInfo(host string, port int, username string) (map[string]string, error) {
	log.Printf("📊 원격 서버 정보 조회: %s@%s:%d", username, host, port)

	// 서버 정보 조회 명령
	sshCommand := `echo "OS: $(uname -s)" && echo "Kernel: $(uname -r)" && echo "Architecture: $(uname -m)" && echo "Hostname: $(hostname)" && echo "Uptime: $(uptime | cut -d',' -f1)" && echo "SSH_KEY_INFO_END"`

	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", username, host),
		sshCommand,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("서버 정보 조회 실패: %v", err)
	}

	// 출력 파싱
	info := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				info[key] = value
			}
		}
	}

	log.Printf("✅ 서버 정보 조회 성공")
	return info, nil
}

// ValidateRemoteServerAccess는 원격 서버 접근 권한을 검증합니다.
func ValidateRemoteServerAccess(host string, port int, username string) error {
	log.Printf("🔐 원격 서버 접근 권한 검증: %s@%s:%d", username, host, port)

	// 기본 연결 테스트
	if err := TestRemoteServerConnection(host, port, username); err != nil {
		return fmt.Errorf("기본 연결 실패: %v", err)
	}

	// SSH 디렉토리 접근 권한 확인
	sshCommand := `test -d ~/.ssh && echo "SSH_DIR_OK" || echo "SSH_DIR_NOT_FOUND"`

	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", username, host),
		sshCommand,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("디렉토리 접근 권한 확인 실패: %v", err)
	}

	if !strings.Contains(string(output), "SSH_DIR_OK") {
		log.Printf("⚠️ ~/.ssh 디렉토리가 존재하지 않습니다. 키 배포시 자동 생성됩니다.")
	}

	log.Printf("✅ 원격 서버 접근 권한 검증 완료")
	return nil
}

// BatchDeployToMultipleServers는 여러 서버에 동시에 키를 배포합니다.
func BatchDeployToMultipleServers(publicKey string, servers []types.ServerDeployTarget) []types.ServerDeployResult {
	log.Printf("🚀 배치 키 배포 시작 (서버 수: %d)", len(servers))

	results := make([]types.ServerDeployResult, len(servers))

	// 고루틴을 사용한 병렬 배포
	resultChan := make(chan types.ServerDeployResult, len(servers))

	for i, server := range servers {
		go func(idx int, srv types.ServerDeployTarget) {
			result := types.ServerDeployResult{
				Index:      idx,
				ServerName: srv.Name,
				Host:       srv.Host,
				Port:       srv.Port,
				Username:   srv.Username,
			}

			startTime := time.Now()
			err := DeploySSHKeyToRemoteServer(publicKey, srv.Host, srv.Port, srv.Username)
			result.Duration = time.Since(startTime)

			if err != nil {
				result.Success = false
				result.ErrorMessage = err.Error()
				log.Printf("❌ 배포 실패 [%s]: %v", srv.Name, err)
			} else {
				result.Success = true
				log.Printf("✅ 배포 성공 [%s]: %v", srv.Name, result.Duration)
			}

			resultChan <- result
		}(i, server)
	}

	// 결과 수집
	for i := 0; i < len(servers); i++ {
		result := <-resultChan
		results[result.Index] = result
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	log.Printf("🎯 배치 키 배포 완료: 성공 %d/%d", successCount, len(servers))
	return results
}
