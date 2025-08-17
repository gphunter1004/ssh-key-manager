package models

import (
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

// UserRole은 사용자 권한을 정의하는 열거형입니다.
type UserRole string

const (
	RoleUser  UserRole = "user"  // 일반 사용자
	RoleAdmin UserRole = "admin" // 관리자
)

// User는 사용자 정보를 저장하는 모델입니다.
type User struct {
	gorm.Model
	Username string   `gorm:"unique;not null"`
	Password string   `gorm:"not null"`
	Role     UserRole `gorm:"not null;default:'user'"`

	// 부서 관련 필드 추가
	DepartmentID *uint      `gorm:"index"`           // 부서 ID (NULL 가능)
	EmployeeID   string     `gorm:"unique;size:20"`  // 사번 (선택사항)
	Position     string     `gorm:"size:50"`         // 직책 (예: 팀장, 사원)
	JoinDate     *time.Time `gorm:""`                // 입사일
	Email        string     `gorm:"unique;size:100"` // 이메일
	Phone        string     `gorm:"size:20"`         // 연락처

	// 관계 정의
	Department *Department `gorm:"foreignKey:DepartmentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"` // 소속 부서
	SSHKeys    []SSHKey    `gorm:"foreignKey:UserID"`                                                     // SSH 키들
	Servers    []Server    `gorm:"foreignKey:UserID"`                                                     // 등록한 서버들

	// 부서 변경 이력
	DepartmentHistories []DepartmentHistory `gorm:"foreignKey:UserID"`
	ChangedHistories    []DepartmentHistory `gorm:"foreignKey:ChangedBy"`
}

// SSHKey는 SSH 키 정보를 저장하는 모델입니다.
type SSHKey struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index"`                                // 인덱스 추가로 성능 향상
	Algorithm  string `gorm:"not null;default:'RSA'"`                        // 알고리즘 (RSA, ECDSA 등)
	Bits       int    `gorm:"not null;default:4096"`                         // 키 크기
	PrivateKey string `gorm:"type:text;not null"`                            // PEM 형식 개인키
	PublicKey  string `gorm:"type:text;not null"`                            // SSH 공개키 (authorized_keys 형식)
	PEM        string `gorm:"type:text;not null"`                            // PEM 형식 개인키 (PrivateKey와 동일)
	PPK        string `gorm:"type:text;not null"`                            // PuTTY 형식 개인키
	User       User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // 외래키 제약조건 추가
}

// Server는 원격 서버 정보를 저장하는 모델입니다.
type Server struct {
	gorm.Model
	UserID      uint   `gorm:"not null;index"`                                // 서버를 등록한 사용자 ID
	Name        string `gorm:"not null"`                                      // 서버 이름 (별칭)
	Host        string `gorm:"not null"`                                      // 서버 IP 또는 호스트명
	Port        int    `gorm:"not null;default:22"`                           // SSH 포트 (기본: 22)
	Username    string `gorm:"not null"`                                      // SSH 접속 계정
	Description string `gorm:"type:text"`                                     // 서버 설명
	Status      string `gorm:"not null;default:'active'"`                     // 서버 상태 (active, inactive)
	User        User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // 외래키 제약조건
}

// ServerKeyDeployment는 서버별 키 배포 기록을 저장하는 모델입니다.
type ServerKeyDeployment struct {
	gorm.Model
	ServerID   uint            `gorm:"not null;index"`                                  // 서버 ID
	SSHKeyID   uint            `gorm:"not null;index"`                                  // SSH 키 ID
	UserID     uint            `gorm:"not null;index"`                                  // 사용자 ID
	Status     string          `gorm:"not null;default:'pending'"`                      // 배포 상태 (pending, success, failed)
	DeployedAt *gorm.DeletedAt `gorm:"index"`                                           // 배포 완료 시간
	ErrorMsg   string          `gorm:"type:text"`                                       // 오류 메시지 (실패시)
	Server     Server          `gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE"` // 외래키 제약조건
	SSHKey     SSHKey          `gorm:"foreignKey:SSHKeyID;constraint:OnDelete:CASCADE"` // 외래키 제약조건
	User       User            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`   // 외래키 제약조건
}
