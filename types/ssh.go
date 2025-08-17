package types

import (
	"ssh-key-manager/models"
	"time"
)

// === SSH 키 관련 ===

// SSHKeyResponse는 API 응답용 SSH 키 정보입니다.
type SSHKeyResponse struct {
	ID          uint      `json:"id"`
	Algorithm   string    `json:"algorithm"`
	Bits        int       `json:"bits"`
	PublicKey   string    `json:"public_key"`
	PEM         string    `json:"pem"`
	PPK         string    `json:"ppk"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Fingerprint string    `json:"fingerprint,omitempty"`
}

// SSHKeyCreateRequest는 SSH 키 생성 요청 구조체입니다.
type SSHKeyCreateRequest struct {
	Algorithm string `json:"algorithm,omitempty"` // RSA, ECDSA 등
	Bits      int    `json:"bits,omitempty"`      // 키 크기
	Comment   string `json:"comment,omitempty"`   // 키 코멘트
}

// SSHKeyListRequest는 SSH 키 목록 요청 구조체입니다.
type SSHKeyListRequest struct {
	PaginationRequest
	SortRequest
	Algorithm string     `json:"algorithm" query:"algorithm"`
	MinBits   int        `json:"min_bits" query:"min_bits"`
	MaxBits   int        `json:"max_bits" query:"max_bits"`
	DateFrom  *time.Time `json:"date_from" query:"date_from"`
	DateTo    *time.Time `json:"date_to" query:"date_to"`
}

// SSHKeyGenerationOptions는 키 생성 옵션 구조체입니다.
type SSHKeyGenerationOptions struct {
	Algorithm    string `json:"algorithm"`     // 알고리즘 (RSA, ECDSA, Ed25519)
	Bits         int    `json:"bits"`          // 키 크기
	Comment      string `json:"comment"`       // 키 코멘트
	Passphrase   string `json:"passphrase"`    // 개인키 암호화 (선택사항)
	ForceReplace bool   `json:"force_replace"` // 기존 키 강제 교체
}

// SSHKeyExportRequest는 키 내보내기 요청 구조체입니다.
type SSHKeyExportRequest struct {
	Format   string `json:"format" binding:"required"` // PEM, PPK, OpenSSH
	Password string `json:"password,omitempty"`        // 암호화 비밀번호 (선택사항)
}

// SSHKeyExportResponse는 키 내보내기 응답 구조체입니다.
type SSHKeyExportResponse struct {
	Format    string `json:"format"`
	Content   string `json:"content"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	Encrypted bool   `json:"encrypted"`
}

// SSHKeyImportRequest는 키 가져오기 요청 구조체입니다.
type SSHKeyImportRequest struct {
	Content   string `json:"content" binding:"required"` // 키 내용
	Format    string `json:"format" binding:"required"`  // 키 형식
	Password  string `json:"password,omitempty"`         // 복호화 비밀번호
	Comment   string `json:"comment,omitempty"`          // 새 코멘트
	Overwrite bool   `json:"overwrite"`                  // 기존 키 덮어쓰기
}

// SSHKeyValidationResult는 키 검증 결과입니다.
type SSHKeyValidationResult struct {
	Valid       bool                   `json:"valid"`
	Algorithm   string                 `json:"algorithm,omitempty"`
	Bits        int                    `json:"bits,omitempty"`
	Fingerprint string                 `json:"fingerprint,omitempty"`
	Comment     string                 `json:"comment,omitempty"`
	Errors      []string               `json:"errors,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	KeyInfo     map[string]interface{} `json:"key_info,omitempty"`
}

// SSHKeyUsageStats는 키 사용 통계입니다.
type SSHKeyUsageStats struct {
	KeyID            uint      `json:"key_id"`
	TotalDeployments int       `json:"total_deployments"`
	ActiveServers    int       `json:"active_servers"`
	LastUsed         time.Time `json:"last_used"`
	FirstUsed        time.Time `json:"first_used"`
	DeploymentRate   float64   `json:"deployment_rate"` // 배포 성공률
}

// SSHKeySecurityCheck는 키 보안 검사 결과입니다.
type SSHKeySecurityCheck struct {
	KeyID           uint     `json:"key_id"`
	SecurityLevel   string   `json:"security_level"` // HIGH, MEDIUM, LOW
	Recommendations []string `json:"recommendations"`
	Issues          []string `json:"issues"`
	Score           int      `json:"score"` // 0-100 점수
}

// === 변환 헬퍼 함수들 ===

// ToSSHKeyResponse는 모델을 SSHKeyResponse로 변환합니다.
func ToSSHKeyResponse(sshKey models.SSHKey, fingerprint string) SSHKeyResponse {
	return SSHKeyResponse{
		ID:          sshKey.ID,
		Algorithm:   sshKey.Algorithm,
		Bits:        sshKey.Bits,
		PublicKey:   sshKey.PublicKey,
		PEM:         sshKey.PEM,
		PPK:         sshKey.PPK,
		CreatedAt:   sshKey.CreatedAt,
		UpdatedAt:   sshKey.UpdatedAt,
		Fingerprint: fingerprint,
	}
}
