package database

import (
	"fmt"
	"log"
	"ssh-key-manager/config"
	"ssh-key-manager/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MigrationManager는 데이터베이스 마이그레이션을 관리합니다.
type MigrationManager struct {
	DB     *gorm.DB
	Config *config.Config
}

// NewMigrationManager는 새로운 마이그레이션 관리자를 생성합니다.
func NewMigrationManager(cfg *config.Config) (*MigrationManager, error) {
	db, err := connectDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	return &MigrationManager{
		DB:     db,
		Config: cfg,
	}, nil
}

// connectDatabase는 데이터베이스에 연결합니다.
func connectDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Printf("✅ 데이터베이스 연결 성공: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	return db, nil
}

// RunMigrations는 모든 필요한 마이그레이션을 실행합니다.
func (m *MigrationManager) RunMigrations() error {
	log.Printf("📦 데이터베이스 마이그레이션 시작...")

	// 테이블 마이그레이션 실행
	if err := m.migrateTables(); err != nil {
		return fmt.Errorf("테이블 마이그레이션 실패: %w", err)
	}

	// 인덱스 생성
	if err := m.createIndexes(); err != nil {
		log.Printf("⚠️ 인덱스 생성 실패 (계속 진행): %v", err)
	}

	// 외래키 제약조건 확인
	if err := m.validateConstraints(); err != nil {
		log.Printf("⚠️ 제약조건 검증 실패 (계속 진행): %v", err)
	}

	// 초기 관리자 계정 생성
	if err := m.createInitialAdmin(); err != nil {
		log.Printf("⚠️ 초기 관리자 계정 생성 실패: %v", err)
	}

	log.Printf("✅ 데이터베이스 마이그레이션 완료")
	return nil
}

// migrateTables는 모든 테이블을 마이그레이션합니다.
func (m *MigrationManager) migrateTables() error {
	models := []interface{}{
		&models.User{},
		&models.SSHKey{},
		&models.Server{},
		&models.ServerKeyDeployment{},
		&models.Department{},
		&models.DepartmentHistory{},
	}

	for _, model := range models {
		if err := m.DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("모델 마이그레이션 실패 %T: %w", model, err)
		}
		log.Printf("   - %T 마이그레이션 완료", model)
	}

	return nil
}

// createIndexes는 성능 향상을 위한 인덱스를 생성합니다.
func (m *MigrationManager) createIndexes() error {
	indexes := []struct {
		table   string
		index   string
		columns []string
	}{
		{"users", "idx_users_username", []string{"username"}},
		{"users", "idx_users_role", []string{"role"}},
		{"users", "idx_users_department_id", []string{"department_id"}},
		{"ssh_keys", "idx_ssh_keys_user_id", []string{"user_id"}},
		{"servers", "idx_servers_user_id", []string{"user_id"}},
		{"servers", "idx_servers_host_port", []string{"host", "port"}},
		{"server_key_deployments", "idx_deployments_server_id", []string{"server_id"}},
		{"server_key_deployments", "idx_deployments_user_id", []string{"user_id"}},
		{"departments", "idx_departments_code", []string{"code"}},
		{"departments", "idx_departments_parent_id", []string{"parent_id"}},
		{"department_histories", "idx_dept_history_user_id", []string{"user_id"}},
		{"department_histories", "idx_dept_history_change_date", []string{"change_date"}},
	}

	for _, idx := range indexes {
		if err := m.createIndexIfNotExists(idx.table, idx.index, idx.columns); err != nil {
			log.Printf("⚠️ 인덱스 생성 실패 %s.%s: %v", idx.table, idx.index, err)
		} else {
			log.Printf("   - 인덱스 생성: %s.%s", idx.table, idx.index)
		}
	}

	return nil
}

// createIndexIfNotExists는 인덱스가 존재하지 않으면 생성합니다.
func (m *MigrationManager) createIndexIfNotExists(table, indexName string, columns []string) error {
	// PostgreSQL에서 인덱스 존재 여부 확인
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = ? AND indexname = ?
		)
	`
	if err := m.DB.Raw(query, table, indexName).Scan(&exists).Error; err != nil {
		return err
	}

	if exists {
		return nil // 이미 존재함
	}

	// 인덱스 생성
	columnsStr := fmt.Sprintf("(%s)", fmt.Sprintf("\"%s\"", columns[0]))
	if len(columns) > 1 {
		var quoted []string
		for _, col := range columns {
			quoted = append(quoted, fmt.Sprintf("\"%s\"", col))
		}
		columnsStr = fmt.Sprintf("(%s)", fmt.Sprintf("%v", quoted))
	}

	createSQL := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s %s",
		indexName, table, columnsStr)

	return m.DB.Exec(createSQL).Error
}

// validateConstraints는 외래키 제약조건을 검증합니다.
func (m *MigrationManager) validateConstraints() error {
	// 외래키 제약조건 확인 쿼리들
	constraints := []struct {
		name  string
		query string
	}{
		{
			"SSH Keys - User 관계",
			"SELECT COUNT(*) FROM ssh_keys sk LEFT JOIN users u ON sk.user_id = u.id WHERE u.id IS NULL",
		},
		{
			"Servers - User 관계",
			"SELECT COUNT(*) FROM servers s LEFT JOIN users u ON s.user_id = u.id WHERE u.id IS NULL",
		},
		{
			"Deployments - Server 관계",
			"SELECT COUNT(*) FROM server_key_deployments d LEFT JOIN servers s ON d.server_id = s.id WHERE s.id IS NULL",
		},
		{
			"Users - Department 관계",
			"SELECT COUNT(*) FROM users u LEFT JOIN departments d ON u.department_id = d.id WHERE u.department_id IS NOT NULL AND d.id IS NULL",
		},
	}

	for _, constraint := range constraints {
		var count int64
		if err := m.DB.Raw(constraint.query).Scan(&count).Error; err != nil {
			log.Printf("⚠️ 제약조건 확인 실패 [%s]: %v", constraint.name, err)
			continue
		}

		if count > 0 {
			log.Printf("⚠️ 제약조건 위반 발견 [%s]: %d건", constraint.name, count)
		} else {
			log.Printf("   - 제약조건 정상 [%s]", constraint.name)
		}
	}

	return nil
}

// CheckConnection은 데이터베이스 연결을 확인합니다.
func (m *MigrationManager) CheckConnection() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// Close는 데이터베이스 연결을 닫습니다.
func (m *MigrationManager) Close() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDatabaseStats는 데이터베이스 통계를 반환합니다.
func (m *MigrationManager) GetDatabaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 테이블별 레코드 수
	tables := []string{"users", "ssh_keys", "servers", "server_key_deployments", "departments", "department_histories"}

	for _, table := range tables {
		var count int64
		if err := m.DB.Table(table).Count(&count).Error; err != nil {
			log.Printf("⚠️ 테이블 카운트 실패 [%s]: %v", table, err)
			continue
		}
		stats[table+"_count"] = count
	}

	// 데이터베이스 버전
	var version string
	if err := m.DB.Raw("SELECT version()").Scan(&version).Error; err == nil {
		stats["database_version"] = version
	}

	return stats, nil
}

// createInitialAdmin은 초기 관리자 계정을 생성합니다.
func (m *MigrationManager) createInitialAdmin() error {
	// 환경변수에서 관리자 계정 정보 확인
	adminUsername := m.Config.AdminUsername
	adminPassword := m.Config.AdminPassword

	if adminUsername == "" || adminPassword == "" {
		log.Printf("📋 관리자 계정 설정이 없습니다. 건너뜀")
		return nil
	}

	log.Printf("👑 초기 관리자 계정 확인 중...")

	// 이미 관리자가 있는지 확인
	var adminCount int64
	if err := m.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount).Error; err != nil {
		return err
	}

	if adminCount > 0 {
		log.Printf("⚠️ 관리자 계정이 이미 존재합니다. 건너뜀")
		return nil
	}

	// 해당 사용자명이 이미 존재하는지 확인
	var existingUser models.User
	result := m.DB.Where("username = ?", adminUsername).First(&existingUser)
	if result.Error == nil {
		// 사용자가 존재하면 관리자로 승격
		log.Printf("🔄 기존 사용자를 관리자로 승격: %s", adminUsername)
		if err := m.DB.Model(&existingUser).Update("role", models.RoleAdmin).Error; err != nil {
			return err
		}
		log.Printf("✅ 사용자 %s가 관리자로 승격되었습니다", adminUsername)
		return nil
	}

	// 새로운 관리자 계정 생성
	hashedPassword, err := m.hashPassword(adminPassword)
	if err != nil {
		log.Printf("❌ 비밀번호 해싱 실패: %v", err)
		return err
	}

	admin := models.User{
		Username: adminUsername,
		Password: hashedPassword,
		Role:     models.RoleAdmin,
	}

	if err := m.DB.Create(&admin).Error; err != nil {
		log.Printf("❌ 관리자 계정 생성 실패: %v", err)
		return err
	}

	log.Printf("✅ 초기 관리자 계정 생성 완료: %s (ID: %d)", adminUsername, admin.ID)
	log.Printf("🔑 관리자 비밀번호: %s", adminPassword)
	log.Printf("⚠️ 보안을 위해 비밀번호를 변경하세요!")

	return nil
}

// hashPassword는 비밀번호를 해시합니다.
func (m *MigrationManager) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func (m *MigrationManager) RunHealthCheck() error {
	log.Printf("🔍 데이터베이스 헬스체크 시작...")

	// 연결 확인
	if err := m.CheckConnection(); err != nil {
		return fmt.Errorf("연결 실패: %w", err)
	}
	log.Printf("   - 연결 상태: 정상")

	// 필수 테이블 존재 확인
	requiredTables := []string{"users", "ssh_keys", "servers"}
	for _, table := range requiredTables {
		if !m.DB.Migrator().HasTable(table) {
			return fmt.Errorf("필수 테이블 누락: %s", table)
		}
	}
	log.Printf("   - 필수 테이블: 정상")

	// 관리자 계정 존재 확인
	var adminCount int64
	if err := m.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount).Error; err != nil {
		log.Printf("⚠️ 관리자 계정 확인 실패: %v", err)
	} else {
		log.Printf("   - 관리자 계정: %d명", adminCount)
		if adminCount == 0 {
			log.Printf("⚠️ 관리자 계정이 없습니다. 초기 설정이 필요할 수 있습니다.")
		}
	}

	log.Printf("✅ 데이터베이스 헬스체크 완료")
	return nil
}
