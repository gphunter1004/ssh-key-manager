package service

import (
	"log"
	"ssh-key-manager/internal/repository"
)

// Container는 모든 서비스 인스턴스를 보관합니다.
type Container struct {
	Auth       *AuthService
	User       *UserService
	Key        *KeyService
	Server     *ServerService
	Department *DepartmentService
}

// 글로벌 컨테이너
var container *Container

// InitializeServices 모든 서비스를 초기화합니다 (단순화).
func InitializeServices() error {
	log.Printf("🔧 서비스 초기화 시작...")

	// Repository 직접 생성 (인터페이스 제거)
	userRepo, err := repository.NewUserRepository()
	if err != nil {
		return err
	}

	keyRepo, err := repository.NewSSHKeyRepository()
	if err != nil {
		return err
	}

	serverRepo, err := repository.NewServerRepository()
	if err != nil {
		return err
	}

	deptRepo, err := repository.NewDepartmentRepository()
	if err != nil {
		return err
	}

	deployRepo, err := repository.NewDeploymentRepository()
	if err != nil {
		return err
	}

	// 서비스 컨테이너 생성 (직접 의존성 주입)
	container = &Container{
		Auth:       NewAuthService(userRepo),
		User:       NewUserService(userRepo),
		Key:        NewKeyService(keyRepo, deployRepo),
		Server:     NewServerService(serverRepo, keyRepo, deployRepo),
		Department: NewDepartmentService(deptRepo),
	}

	log.Printf("✅ 서비스 초기화 완료")
	return nil
}

// C 서비스 컨테이너를 반환합니다.
func C() *Container {
	if container == nil {
		panic("Service container not initialized. Call InitializeServices() first.")
	}
	return container
}
