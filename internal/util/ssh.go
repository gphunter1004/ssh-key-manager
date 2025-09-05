package util

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// TestSSHConnection은 SSH 연결을 테스트합니다.
func TestSSHConnection(host string, port int, username string) error {
	log.Printf("🔍 SSH 연결 테스트 중: %s@%s:%d", username, host, port)

	// SSH 연결 테스트 명령
	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes", // 비밀번호 프롬프트 비활성화
		"-o", "ConnectTimeout=10", // 연결 타임아웃 10초
		"-o", "StrictHostKeyChecking=no", // 호스트 키 확인 비활성화
		"-o", "UserKnownHostsFile=/dev/null", // known_hosts 파일 사용 안함
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
func DeploySSHKeyToServer(publicKey, host string, port int, username string, targetDirectory string) error {
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
	if err := deployPublicKeyViaSSH(publicKey, host, port, username, targetDirectory); err != nil {
		return fmt.Errorf("공개키 배포 실패: %v", err)
	}

	log.Printf("✅ SSH 키 배포 완료: %s@%s:%d", username, host, port)
	return nil
}

// deployPublicKeyViaSSH는 SSH를 통해 공개키를 원격 서버에 배포합니다.
func deployPublicKeyViaSSH(publicKey, host string, port int, username string, targetDirectory string) error {
	log.Printf("🔑 공개키 배포 중...")

	// 공개키를 정리 (개행 제거 등)
	cleanedKey := strings.TrimSpace(publicKey)
	if !strings.HasSuffix(cleanedKey, "\n") {
		cleanedKey += "\n"
	}

	authorizedKeysPath := fmt.Sprintf("%s/authorized_keys", targetDirectory)

	// SSH를 통해 authorized_keys에 공개키 추가
	// 중복을 방지하기 위해 먼저 키가 있는지 확인 후 추가
	sshCommand := fmt.Sprintf(
		`mkdir -p %s && chmod 700 %s && `+
			`touch %s && `+ // 파일이 없는 경우 생성
			`if ! grep -q "%s" %s 2>/dev/null; then `+
			`echo '%s' >> %s; fi && `+
			`chmod 600 %s && `+
			`echo 'Key deployed successfully'`,
		targetDirectory, targetDirectory, // mkdir, chmod
		authorizedKeysPath,                                // touch
		strings.Fields(cleanedKey)[1], authorizedKeysPath, // grep
		strings.ReplaceAll(cleanedKey, "'", "'\"'\"'"), authorizedKeysPath, // echo
		authorizedKeysPath, // chmod
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
