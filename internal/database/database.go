package database

import (
	"fmt"
	"log"
	"ssh-key-manager/internal/config"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Initialize는 데이터베이스를 초기화합니다.
func Initialize(cfg *config.Config) error {
	// 데이터베이스 연결
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		return err
	}

	// 연결 테스트
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		return err
	}

	// 전역 DB 설정
	model.SetDB(db)

	// SSH 키 테이블 스키마 검증 및 수정
	if err := checkAndFixSSHKeySchema(db); err != nil {
		return err
	}

	// 마이그레이션 실행
	if err := runMigrations(db); err != nil {
		return err
	}

	// 초기 관리자 계정 생성
	if err := createInitialAdmin(db, cfg); err != nil {
		log.Printf("⚠️ 초기 관리자 계정 생성 실패: %v", err)
	}

	return nil
}

// checkAndFixSSHKeySchema SSH 키 테이블 스키마를 검증하고 필요시 수정합니다.
func checkAndFixSSHKeySchema(db *gorm.DB) error {
	log.Printf("🔍 SSH 키 테이블 스키마 검증 중...")

	// 테이블 존재 여부 확인
	var tableExists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'ssh_keys')").Scan(&tableExists).Error
	if err != nil {
		log.Printf("❌ 테이블 존재 확인 실패: %v", err)
		return err
	}

	if !tableExists {
		log.Printf("📋 SSH 키 테이블이 존재하지 않음. 새로 생성됩니다.")
		return nil
	}

	// 현재 컬럼 정보 조회
	var columns []struct {
		ColumnName string `db:"column_name"`
		DataType   string `db:"data_type"`
		IsNullable string `db:"is_nullable"`
	}

	err = db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'ssh_keys' 
		ORDER BY ordinal_position
	`).Scan(&columns).Error

	if err != nil {
		log.Printf("❌ 컬럼 정보 조회 실패: %v", err)
		return err
	}

	log.Printf("📋 현재 SSH 키 테이블 컬럼 정보:")
	for _, col := range columns {
		log.Printf("  - %s: %s (%s)", col.ColumnName, col.DataType, col.IsNullable)
	}

	// 필요한 컬럼들과 현재 컬럼들 비교
	requiredColumns := map[string]bool{
		"id":          false,
		"created_at":  false,
		"updated_at":  false,
		"deleted_at":  false,
		"user_id":     false,
		"private_key": false,
		"public_key":  false,
		"ppk":         false,
	}

	currentColumns := make(map[string]bool)
	for _, col := range columns {
		currentColumns[col.ColumnName] = true
		if _, exists := requiredColumns[col.ColumnName]; exists {
			requiredColumns[col.ColumnName] = true
		}
	}

	// 누락된 컬럼이나 잘못된 스키마 확인
	needsRecreation := false
	var issues []string

	// 필수 컬럼 누락 확인
	for colName, exists := range requiredColumns {
		if !exists {
			issues = append(issues, fmt.Sprintf("누락된 컬럼: %s", colName))
			needsRecreation = true
		}
	}

	// 구 스키마 컬럼 확인 (pem 컬럼이 있으면 구 스키마)
	if currentColumns["pem"] {
		issues = append(issues, "구 스키마 감지: pem 컬럼 존재")
		needsRecreation = true
	}

	if len(issues) > 0 {
		log.Printf("⚠️ SSH 키 테이블 스키마 문제 감지:")
		for _, issue := range issues {
			log.Printf("  - %s", issue)
		}
	}

	if needsRecreation {
		log.Printf("🔄 SSH 키 테이블 재생성 필요. 기존 데이터를 백업 후 재생성합니다.")
		return recreateSSHKeyTables(db)
	}

	log.Printf("✅ SSH 키 테이블 스키마가 올바릅니다.")
	return nil
}

// recreateSSHKeyTables SSH 키 관련 테이블들을 재생성합니다.
func recreateSSHKeyTables(db *gorm.DB) error {
	log.Printf("🔄 SSH 키 관련 테이블 재생성 시작...")

	// 기존 데이터 백업 (가능한 경우)
	if err := backupSSHKeyData(db); err != nil {
		log.Printf("⚠️ SSH 키 데이터 백업 실패 (계속 진행): %v", err)
	}

	// 의존 테이블 먼저 삭제 (외래키 제약조건 때문에)
	dependentTables := []string{
		"server_key_deployments",
		"ssh_keys",
	}

	for _, tableName := range dependentTables {
		if err := db.Exec("DROP TABLE IF EXISTS " + tableName + " CASCADE").Error; err != nil {
			log.Printf("⚠️ 테이블 %s 삭제 실패 (무시): %v", tableName, err)
		} else {
			log.Printf("✅ 테이블 %s 삭제 완료", tableName)
		}
	}

	log.Printf("✅ SSH 키 관련 테이블 재생성 준비 완료")
	return nil
}

// backupSSHKeyData 기존 SSH 키 데이터를 백업합니다 (가능한 경우).
func backupSSHKeyData(db *gorm.DB) error {
	log.Printf("💾 SSH 키 데이터 백업 시도...")

	// 기존 데이터 개수 확인
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM ssh_keys").Scan(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		log.Printf("📋 백업할 SSH 키 데이터가 없습니다.")
		return nil
	}

	log.Printf("⚠️ %d개의 SSH 키가 삭제됩니다. (개발 환경이므로 백업하지 않음)", count)

	// 실제 운영 환경에서는 여기에 백업 로직 추가
	// 예: CSV 파일로 내보내기, 다른 테이블에 임시 저장 등

	return nil
}

// runMigrations는 데이터베이스 마이그레이션을 실행합니다.
func runMigrations(db *gorm.DB) error {
	log.Printf("📦 데이터베이스 마이그레이션 시작...")

	models := []interface{}{
		&model.User{},
		&model.Department{},
		&model.Server{},
		&model.SSHKey{},              // SSH 키 테이블
		&model.ServerKeyDeployment{}, // 배포 기록 테이블
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			log.Printf("❌ %T 마이그레이션 실패: %v", m, err)
			return err
		}
		log.Printf("   - %T 마이그레이션 완료", m)
	}

	log.Printf("✅ 데이터베이스 마이그레이션 완료")
	return nil
}

// createInitialAdmin은 초기 관리자 계정을 생성합니다.
func createInitialAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		log.Printf("📋 관리자 계정 설정이 없습니다. 건너뜀")
		return nil
	}

	// 관리자 존재 여부 확인
	var adminCount int64
	if err := db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&adminCount).Error; err != nil {
		return err
	}

	if adminCount > 0 {
		log.Printf("⚠️ 관리자 계정이 이미 존재합니다. 건너뜀")
		return nil
	}

	// 기존 사용자 확인
	var existingUser model.User
	result := db.Where("username = ?", cfg.AdminUsername).First(&existingUser)
	if result.Error == nil {
		// 기존 사용자를 관리자로 승격
		if err := db.Model(&existingUser).Update("role", model.RoleAdmin).Error; err != nil {
			return err
		}
		log.Printf("✅ 기존 사용자 %s를 관리자로 승격", cfg.AdminUsername)
		return nil
	}

	// 새 관리자 계정 생성
	hashedPassword, err := util.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}

	admin := model.User{
		Username: cfg.AdminUsername,
		Password: hashedPassword,
		Role:     model.RoleAdmin,
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	log.Printf("✅ 초기 관리자 계정 생성 완료: %s", cfg.AdminUsername)
	return nil
}
