package util

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// TestSSHConnection은 SSH 연결을 테스트합니다.
func TestSSHConnection(host string, port int, username string) error {
	log.Printf("🔍 SSH 연결 테스트 중: %s@%s:%d", username, host, port)

	// SSH 연결 테스트 명령
	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes",                    // 비밀번호 프롬프트 비활성화
		"-o", "ConnectTimeout=10",                // 연결 타임아웃 10초
		"-o", "StrictHostKeyChecking=no",         // 호스트 키 확인 비활성화
		"-o", "UserKnownHostsFile=/dev/null",     // known_hosts 파일 사용 안함
		"-o", "PreferredAuthentications=publickey", // 공개키 인증만 시도
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", username, host),
		"echo 'SSH connection test successful'",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ SSH 연결 실패: %s", string(output))
		return fmt.Errorf("SSH 연결 실패 (%s@%s:%d): %v", username, host, port, err)
	}

	if strings.Contains(string(output), "SSH connection test successful") {
		log.Printf("✅ SSH 연결 테스트 성공")
		return nil
	}

	return fmt.Errorf("SSH 연결 테스트 응답 확인 실패: %s", string(output))
}

// DeploySSHKeyToServer는 SSH 키를 원격 서버에 배포합니다.
func DeploySSHKeyToServer(publicKey, host string, port int, username string) error {
	log.Printf("📡 원격 서버 SSH 키 배포 시작: %s@%s:%d", username, host, port)

	// 공개키 검증
	if strings.TrimSpace(publicKey) == "" {
		return fmt.Errorf("공개키가 비어있습니다")
	}

	// 먼저 연결 테스트
	if err := TestSSHConnection(host, port, username); err != nil {
		return fmt.Errorf("SSH 연결 테스트 실패: %v", err)
	}

	// 공개키 배포
	if err := deployPublicKeyViaSSH(publicKey, host, port, username); err != nil {
		return fmt.Errorf("공개키 배포 실패: %v", err)
	}

	log.Printf("✅ SSH 키 배포 완료: %s@%s:%d", username, host, port)
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
	// 중복을 방지하기 위해 먼저 키가 있는지 확인 후 추가
	sshCommand := fmt.Sprintf(
		`mkdir -p ~/.ssh && chmod 700 ~/.ssh && `+
			`if ! grep -q "%s" ~/.ssh/authorized_keys 2>/dev/null; then `+
			`echo '%s' >> ~/.ssh/authorized_keys; fi && `+
			`chmod 600 ~/.ssh/authorized_keys && `+
			`echo 'Key deployed successfully'`,
		strings.Fields(cleanedKey)[1], // 키의 핵심 부분만 추출하여 중복 체크
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

// RemoveSSHKeyFromServer는 원격 서버에서 SSH 키를 제거합니다.
func RemoveSSHKeyFromServer(publicKey, host string, port int, username string) error {
	log.Printf("🗑️ 원격 서버에서 SSH 키 제거 시작: %s@%s:%d", username, host, port)

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
		`if [ -f ~/.ssh/authorized_keys ]; then `+
			`grep -v '%s' ~/.ssh/authorized_keys > ~/.ssh/authorized_keys.tmp && `+
			`mv ~/.ssh/authorized_keys.tmp ~/.ssh/authorized_keys && `+
			`echo 'Key removed successfully'; `+
			`else echo 'No authorized_keys file found'; fi`,
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

// GetServerInfo는 원격 서버의 기본 정보를 조회합니다.
func GetServerInfo(host string, port int, username string) (map[string]string, error) {
	log.Printf("📊 원격 서버 정보 조회: %s@%s:%d", username, host, port)

	// 서버 정보 조회 명령
	sshCommand := `echo "OS: $(uname -s)" && echo "Kernel: $(uname -r)" && echo "Architecture: $(uname -m)" && echo "Hostname: $(hostname)" && echo "Uptime: $(uptime | cut -d',' -f1)" && echo "SSH_INFO_END"`

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
		if strings.Contains(line, ":") && !strings.Contains(line, "SSH_INFO_END") {
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

// ValidateSSHKeyFormat은 SSH 공개키 형식을 검증합니다.
func ValidateSSHKeyFormat(publicKey string) error {
	cleanedKey := strings.TrimSpace(publicKey)
	if cleanedKey == "" {
		return fmt.Errorf("공개키가 비어있습니다")
	}

	parts := strings.Fields(cleanedKey)
	if len(parts) < 2 {
		return fmt.Errorf("유효하지 않은 SSH 공개키 형식입니다")
	}

	// 지원되는 키 타입 확인
	supportedTypes := []string{"ssh-rsa", "ssh-dss", "ssh-ed25519", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521"}
	keyType := parts[0]
	
	supported := false
	for _, supportedType := range supportedTypes {
		if keyType == supportedType {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("지원되지 않는 SSH 키 타입입니다: %s", keyType)
	}

	return nil
}

// BatchDeployToServers는 여러 서버에 동시에 키를 배포합니다.
func BatchDeployToServers(publicKey string, servers []ServerTarget) []DeploymentResult {
	log.Printf("🚀 배치 키 배포 시작 (서버 수: %d)", len(servers))

	results := make([]DeploymentResult, len(servers))
	resultChan := make(chan DeploymentResultWithIndex, len(servers))

	// 고루틴을 사용한 병렬 배포
	for i, server := range servers {
		go func(idx int, srv ServerTarget) {
			result := DeploymentResultWithIndex{
				Index: idx,
				Result: DeploymentResult{
					ServerName: srv.Name,
					Host:       srv.Host,
					Port:       srv.Port,
					Username:   srv.Username,
				},
			}

			startTime := time.Now()
			err := DeploySSHKeyToServer(publicKey, srv.Host, srv.Port, srv.Username)
			result.Result.Duration = time.Since(startTime)

			if err != nil {
				result.Result.Success = false
				result.Result.ErrorMessage = err.Error()
				log.Printf("❌ 배포 실패 [%s]: %v", srv.Name, err)
			} else {
				result.Result.Success = true
				log.Printf("✅ 배포 성공 [%s]: %v", srv.Name, result.Result.Duration)
			}

			resultChan <- result
		}(i, server)
	}

	// 결과 수집
	for i := 0; i < len(servers); i++ {
		result := <-resultChan
		results[result.Index] = result.Result
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

// 배치 배포를 위한 헬퍼 구조체들
type ServerTarget struct {
	Name     string
	Host     string
	Port     int
	Username string
}

type DeploymentResult struct {
	ServerName   string        `json:"server_name"`
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Username     string        `json:"username"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Duration     time.Duration `json:"duration"`
}

type DeploymentResultWithIndex struct {
	Index  int
	Result DeploymentResult
}
