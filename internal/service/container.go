// internal/service/container.go
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

// InitializeServices 모든 서비스를 초기화합니다.
func InitializeServices() error {
	log.Printf("🔧 서비스 초기화 시작...")

	// Repository 생성
	repos, err := repository.NewRepositories()
	if err != nil {
		return err
	}

	// 서비스 컨테이너 생성
	container = &Container{
		Auth:       NewAuthService(repos),
		User:       NewUserService(repos),
		Key:        NewKeyService(repos),
		Server:     NewServerService(repos),
		Department: NewDepartmentService(repos),
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
