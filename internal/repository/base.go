package repository

import (
	"ssh-key-manager/internal/model"

	"gorm.io/gorm"
)

// ========== Repository Interfaces ==========

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

// ========== Base Repository ==========

// BaseRepository 모든 Repository의 기본 구조체
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 기본 Repository 생성자
func NewBaseRepository() (*BaseRepository, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, err
	}
	return &BaseRepository{db: db}, nil
}

// GetDB 데이터베이스 인스턴스 반환
func (br *BaseRepository) GetDB() *gorm.DB {
	return br.db
}

// ========== Transaction Manager ==========

// TransactionManagerImpl 트랜잭션 매니저 구현체
type TransactionManagerImpl struct {
	db *gorm.DB
}

// NewTransactionManager 트랜잭션 매니저 생성자
func NewTransactionManager() (TransactionManager, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, err
	}
	return &TransactionManagerImpl{db: db}, nil
}

// WithTransaction 트랜잭션 실행
func (tm *TransactionManagerImpl) WithTransaction(fn func(*gorm.DB) error) error {
	return tm.db.Transaction(fn)
}

// ========== Repository Container ==========

// Repositories 모든 Repository를 포함하는 구조체
type Repositories struct {
	User       UserRepository
	SSHKey     SSHKeyRepository
	Server     ServerRepository
	Department DepartmentRepository
	Deployment DeploymentRepository
	TxManager  TransactionManager
}

// NewRepositories 모든 Repository 인스턴스 생성
func NewRepositories() (*Repositories, error) {
	userRepo, err := NewUserRepository()
	if err != nil {
		return nil, err
	}

	keyRepo, err := NewSSHKeyRepository()
	if err != nil {
		return nil, err
	}

	serverRepo, err := NewServerRepository()
	if err != nil {
		return nil, err
	}

	deptRepo, err := NewDepartmentRepository()
	if err != nil {
		return nil, err
	}

	deployRepo, err := NewDeploymentRepository()
	if err != nil {
		return nil, err
	}

	txManager, err := NewTransactionManager()
	if err != nil {
		return nil, err
	}

	return &Repositories{
		User:       userRepo,
		SSHKey:     keyRepo,
		Server:     serverRepo,
		Department: deptRepo,
		Deployment: deployRepo,
		TxManager:  txManager,
	}, nil
}
