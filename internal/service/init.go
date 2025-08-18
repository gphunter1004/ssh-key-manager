package service

import (
	"log"
	"ssh-key-manager/internal/repository"
)

// 글로벌 서비스 인스턴스들
var (
	globalRepos             *repository.Repositories
	globalAuthService       *AuthService
	globalUserService       *UserService
	globalKeyService        *KeyService
	globalServerService     *ServerService
	globalDepartmentService *DepartmentService
)

// InitializeServices 모든 서비스를 초기화합니다.
func InitializeServices() error {
	log.Printf("🔧 서비스 초기화 시작...")

	// Repository 생성
	repos, err := repository.NewRepositories()
	if err != nil {
		return err
	}
	globalRepos = repos

	// 각 서비스 초기화
	globalAuthService = NewAuthService(repos)
	globalUserService = NewUserService(repos)
	globalKeyService = NewKeyService(repos)
	globalServerService = NewServerService(repos)
	globalDepartmentService = NewDepartmentService(repos)

	log.Printf("✅ 서비스 초기화 완료")
	return nil
}

// GetRepositories Repository 인스턴스 반환
func GetRepositories() *repository.Repositories {
	return globalRepos
}

// ========== 서비스 접근자 함수들 ==========

// GetAuthService 인증 서비스 인스턴스 반환
func GetAuthService() *AuthService {
	return globalAuthService
}

// GetUserService 사용자 서비스 인스턴스 반환
func GetUserService() *UserService {
	return globalUserService
}

// GetKeyService 키 서비스 인스턴스 반환
func GetKeyService() *KeyService {
	return globalKeyService
}

// GetServerService 서버 서비스 인스턴스 반환
func GetServerService() *ServerService {
	return globalServerService
}

// GetDepartmentService 부서 서비스 인스턴스 반환
func GetDepartmentService() *DepartmentService {
	return globalDepartmentService
}
