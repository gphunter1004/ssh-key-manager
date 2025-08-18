package repository

import (
	"ssh-key-manager/internal/model"

	"gorm.io/gorm"
)

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
