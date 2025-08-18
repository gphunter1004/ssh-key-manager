package repository

import (
	"ssh-key-manager/internal/model"

	"gorm.io/gorm"
)

// UserRepository 사용자 데이터 접근 인터페이스
type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	FindAll() ([]model.User, error)
	ExistsByID(id uint) (bool, error)
	ExistsByUsername(username string) (bool, error)
	CountByRole(role model.UserRole) (int64, error)
}

// SSHKeyRepository SSH 키 데이터 접근 인터페이스
type SSHKeyRepository interface {
	Create(key *model.SSHKey) error
	FindByUserID(userID uint) (*model.SSHKey, error)
	DeleteByUserID(userID uint) error
	ExistsByUserID(userID uint) (bool, error)
	ReplaceUserKey(userID uint, key *model.SSHKey) error
	GetStatistics() (map[string]interface{}, error)
}

// ServerRepository 서버 데이터 접근 인터페이스
type ServerRepository interface {
	Create(server *model.Server) error
	FindByID(id uint) (*model.Server, error)
	FindByUserID(userID uint) ([]model.Server, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	ExistsByUserAndHost(userID uint, host string, port int) (bool, error)
}

// DepartmentRepository 부서 데이터 접근 인터페이스
type DepartmentRepository interface {
	Create(dept *model.Department) error
	FindByID(id uint) (*model.Department, error)
	FindAll(includeInactive bool) ([]model.Department, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	ExistsByCode(code string) (bool, error)
	FindChildren(parentID uint) ([]model.Department, error)
	CountUsers(deptID uint) (int64, error)
	FindUsersByDepartment(deptID uint) ([]model.User, error)
}

// DeploymentRepository 배포 기록 데이터 접근 인터페이스
type DeploymentRepository interface {
	Create(deployment *model.ServerKeyDeployment) error
	FindByUserID(userID uint) ([]model.ServerKeyDeployment, error)
	DeleteBySSHKeyID(keyID uint) error
	DeleteByServerID(serverID uint) error
}

// TransactionManager 트랜잭션 관리 인터페이스
type TransactionManager interface {
	WithTransaction(fn func(*gorm.DB) error) error
}

// Repositories 모든 Repository를 포함하는 구조체
type Repositories struct {
	User       UserRepository
	SSHKey     SSHKeyRepository
	Server     ServerRepository
	Department DepartmentRepository
	Deployment DeploymentRepository
	TxManager  TransactionManager
}
