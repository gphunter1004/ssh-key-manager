package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB

// SetDB는 전역 DB 인스턴스를 설정합니다.
func SetDB(db *gorm.DB) {
	DB = db
}

// SafeDB는 안전한 DB 인스턴스를 반환합니다.
func SafeDB() *gorm.DB {
	if DB == nil {
		panic("Database not initialized. Call SetDB() first.")
	}
	return DB
}

// IsDBInitialized는 DB가 초기화되었는지 확인합니다.
func IsDBInitialized() bool {
	return DB != nil
}

// GetDB는 DB 인스턴스를 안전하게 반환합니다 (에러 반환 버전).
func GetDB() (*gorm.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return DB, nil
}

// UserRole은 사용자 권한을 정의합니다.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User는 사용자 정보를 저장하는 모델입니다.
type User struct {
	gorm.Model
	Username string   `gorm:"unique;not null" json:"username"`
	Password string   `gorm:"not null" json:"-"`
	Role     UserRole `gorm:"not null;default:'user'" json:"role"`

	// 부서 관련 필드
	DepartmentID *uint      `gorm:"index" json:"department_id"`
	EmployeeID   string     `gorm:"unique;size:20" json:"employee_id"`
	Position     string     `gorm:"size:50" json:"position"`
	JoinDate     *time.Time `json:"join_date"`
	Email        string     `gorm:"unique;size:100" json:"email"`
	Phone        string     `gorm:"size:20" json:"phone"`

	// 관계
	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	SSHKeys    []SSHKey    `gorm:"foreignKey:UserID" json:"ssh_keys,omitempty"`
	Servers    []Server    `gorm:"foreignKey:UserID" json:"servers,omitempty"`
}

// SSHKey는 SSH 키 정보를 저장하는 모델입니다.
type SSHKey struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	Algorithm  string `gorm:"not null;default:'RSA'" json:"algorithm"`
	Bits       int    `gorm:"not null;default:4096" json:"bits"`
	PrivateKey string `gorm:"type:text;not null" json:"private_key"`
	PublicKey  string `gorm:"type:text;not null" json:"public_key"`
	PEM        string `gorm:"type:text;not null" json:"pem"`
	PPK        string `gorm:"type:text;not null" json:"ppk"`

	// 관계
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// Server는 원격 서버 정보를 저장하는 모델입니다.
type Server struct {
	gorm.Model
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	Name        string `gorm:"not null" json:"name"`
	Host        string `gorm:"not null" json:"host"`
	Port        int    `gorm:"not null;default:22" json:"port"`
	Username    string `gorm:"not null" json:"username"`
	Description string `gorm:"type:text" json:"description"`
	Status      string `gorm:"not null;default:'active'" json:"status"`

	// 관계
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// ServerKeyDeployment는 서버별 키 배포 기록을 저장하는 모델입니다.
type ServerKeyDeployment struct {
	gorm.Model
	ServerID   uint            `gorm:"not null;index" json:"server_id"`
	SSHKeyID   uint            `gorm:"not null;index" json:"ssh_key_id"`
	UserID     uint            `gorm:"not null;index" json:"user_id"`
	Status     string          `gorm:"not null;default:'pending'" json:"status"`
	DeployedAt *gorm.DeletedAt `gorm:"index" json:"deployed_at"`
	ErrorMsg   string          `gorm:"type:text" json:"error_msg"`

	// 관계
	Server Server `gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE" json:"server,omitempty"`
	SSHKey SSHKey `gorm:"foreignKey:SSHKeyID;constraint:OnDelete:CASCADE" json:"ssh_key,omitempty"`
	User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// Department는 부서 정보를 저장하는 모델입니다.
type Department struct {
	gorm.Model
	Code        string `gorm:"unique;not null;index" json:"code"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ParentID    *uint  `gorm:"index" json:"parent_id"`
	Level       int    `gorm:"not null;default:1" json:"level"`
	IsActive    bool   `gorm:"not null;default:true" json:"is_active"`

	// 관계
	Parent   *Department  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Users    []User       `gorm:"foreignKey:DepartmentID" json:"users,omitempty"`
}
