package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config는 애플리케이션의 모든 설정을 담는 구조체입니다.
type Config struct {
	// Database settings
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT settings
	JWTSecret string

	// Server settings
	ServerPort string
	KeyBits    int

	// SSH Key Auto Installation settings
	AutoInstallKeys bool
	SSHUser         string
	SSHHomePath     string

	// Admin settings
	AdminUsername string
	AdminPassword string
}

// envDefaults는 .env 파일 생성 시 사용할 기본값들을 정의합니다.
var envDefaults = map[string]string{
	"DB_HOST":           "postgres",
	"DB_PORT":           "5432",
	"DB_USER":           "postgres",
	"DB_PASSWORD":       "password",
	"DB_NAME":           "key-manager",
	"JWT_SECRET":        "", // 런타임에 생성됨
	"SERVER_PORT":       "8080",
	"KEY_BITS":          "4096",
	"AUTO_INSTALL_KEYS": "false",
	"SSH_USER":          "robos",
	"SSH_HOME_PATH":     "/home/$SSH_USER",
	"ADMIN_USERNAME":    "admin",
	"ADMIN_PASSWORD":    "", // 런타임에 생성됨
}

// LoadConfig는 .env 파일에서 모든 설정을 로드합니다.
func LoadConfig() (*Config, error) {
	envFilePath := ".env"

	// .env 파일 존재 여부 확인
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		log.Printf("정보: .env 파일이 존재하지 않습니다. 새로 생성합니다.")
		if err := createEnvFile(envFilePath); err != nil {
			log.Printf("경고: .env 파일 생성 실패: %v", err)
		} else {
			log.Printf("정보: .env 파일이 성공적으로 생성되었습니다: %s", envFilePath)
		}
	}

	// .env 파일 로드
	if err := godotenv.Load(envFilePath); err != nil {
		log.Printf("경고: .env 파일을 로드할 수 없습니다. 환경 변수를 직접 사용합니다. 오류: %v", err)
	}

	cfg := &Config{
		DBHost:        getEnv("DB_HOST", envDefaults["DB_HOST"]),
		DBPort:        getEnv("DB_PORT", envDefaults["DB_PORT"]),
		DBUser:        getEnv("DB_USER", envDefaults["DB_USER"]),
		DBPassword:    getEnv("DB_PASSWORD", envDefaults["DB_PASSWORD"]),
		DBName:        getEnv("DB_NAME", envDefaults["DB_NAME"]),
		JWTSecret:     getEnv("JWT_SECRET", envDefaults["JWT_SECRET"]),
		ServerPort:    getEnv("SERVER_PORT", envDefaults["SERVER_PORT"]),
		SSHUser:       getEnv("SSH_USER", envDefaults["SSH_USER"]),
		SSHHomePath:   expandEnvVars(getEnv("SSH_HOME_PATH", envDefaults["SSH_HOME_PATH"])),
		AdminUsername: getEnv("ADMIN_USERNAME", envDefaults["ADMIN_USERNAME"]),
		AdminPassword: getEnv("ADMIN_PASSWORD", envDefaults["ADMIN_PASSWORD"]),
	}

	// KeyBits 파싱 (정수형)
	keyBitsStr := getEnv("KEY_BITS", envDefaults["KEY_BITS"])
	keyBits, err := strconv.Atoi(keyBitsStr)
	if err != nil {
		log.Printf("경고: KEY_BITS 파싱 실패, 기본값(4096) 사용. 오류: %v", err)
		cfg.KeyBits = 4096 // 파싱 실패 시 기본값
	} else {
		cfg.KeyBits = keyBits
	}

	// AutoInstallKeys 파싱 (불리언)
	autoInstallStr := getEnv("AUTO_INSTALL_KEYS", envDefaults["AUTO_INSTALL_KEYS"])
	autoInstall, err := strconv.ParseBool(autoInstallStr)
	if err != nil {
		log.Printf("경고: AUTO_INSTALL_KEYS 파싱 실패, 기본값(false) 사용. 오류: %v", err)
		cfg.AutoInstallKeys = false // 파싱 실패 시 기본값
	} else {
		cfg.AutoInstallKeys = autoInstall
	}

	// 설정 검증
	if err := validateConfig(cfg); err != nil {
		log.Printf("경고: 설정 검증 실패: %v", err)
	}

	return cfg, nil
}

// createEnvFile은 현재 환경변수와 기본값을 기반으로 .env 파일을 생성합니다.
func createEnvFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("파일 생성 실패: %w", err)
	}
	defer file.Close()

	// 헤더 작성
	fmt.Fprintf(file, "# 환경설정 파일 (자동 생성됨)\n")
	fmt.Fprintf(file, "# 생성 시간: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "# 주의: 이 파일을 수정한 후 애플리케이션을 재시작하세요.\n\n")

	// 데이터베이스 설정
	fmt.Fprintf(file, "# 데이터베이스 설정\n")
	writeEnvVar(file, "DB_HOST", "데이터베이스 호스트")
	writeEnvVar(file, "DB_PORT", "데이터베이스 포트")
	writeEnvVar(file, "DB_USER", "데이터베이스 사용자명")
	writeEnvVar(file, "DB_PASSWORD", "데이터베이스 비밀번호")
	writeEnvVar(file, "DB_NAME", "데이터베이스 이름")

	fmt.Fprintf(file, "\n# JWT 설정\n")
	writeEnvVarWithGenerator(file, "JWT_SECRET", "JWT 서명 키 (자동 생성됨)", generateJWTSecret)

	fmt.Fprintf(file, "\n# 서버 설정\n")
	writeEnvVar(file, "SERVER_PORT", "서버 포트")
	writeEnvVar(file, "KEY_BITS", "SSH 키 비트 수")

	fmt.Fprintf(file, "\n# SSH 키 자동 설치 설정\n")
	writeEnvVar(file, "AUTO_INSTALL_KEYS", "SSH 키 자동 설치 여부")
	writeEnvVar(file, "SSH_USER", "SSH 사용자명")
	writeEnvVar(file, "SSH_HOME_PATH", "SSH 홈 디렉토리 경로")

	fmt.Fprintf(file, "\n# 관리자 설정\n")
	writeEnvVar(file, "ADMIN_USERNAME", "초기 관리자 사용자명")
	writeEnvVar(file, "ADMIN_PASSWORD", "초기 관리자 비밀번호")
	//writeEnvVarWithGenerator(file, "ADMIN_PASSWORD", "초기 관리자 비밀번호 (자동 생성됨)", "")

	// 추가 환경변수가 있다면 포함
	fmt.Fprintf(file, "\n# 추가 환경변수 (컨테이너에서 감지됨)\n")
	for _, env := range os.Environ() {
		if shouldIncludeExtraEnv(env) {
			fmt.Fprintf(file, "%s\n", env)
		}
	}

	return nil
}

// writeEnvVar는 환경변수를 .env 파일에 작성합니다.
func writeEnvVar(file *os.File, key, comment string) {
	value := getEnvValue(key)
	fmt.Fprintf(file, "# %s\n", comment)
	fmt.Fprintf(file, "%s=%s\n\n", key, value)
}

// writeEnvVarWithGenerator는 생성 함수를 사용하여 환경변수를 .env 파일에 작성합니다.
func writeEnvVarWithGenerator(file *os.File, key, comment string, generator func() string) {
	value := getEnvValueWithGenerator(key, generator)
	fmt.Fprintf(file, "# %s\n", comment)
	fmt.Fprintf(file, "%s=%s\n\n", key, value)
}

// getEnvValue는 현재 환경변수 값을 반환하거나, 없으면 기본값을 반환합니다.
func getEnvValue(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if defaultValue, exists := envDefaults[key]; exists {
		return defaultValue
	}
	return ""
}

// getEnvValueWithGenerator는 현재 환경변수 값을 반환하거나, 없으면 생성 함수를 사용합니다.
func getEnvValueWithGenerator(key string, generator func() string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	// 생성 함수가 있으면 사용, 없으면 기본값
	if generator != nil {
		return generator()
	}
	if defaultValue, exists := envDefaults[key]; exists {
		return defaultValue
	}
	return ""
}

// generateJWTSecret은 안전한 JWT 시크릿을 생성합니다.
func generateJWTSecret() string {
	// 256비트 (32바이트) 랜덤 키 생성
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("경고: JWT 시크릿 생성 실패, 기본값 사용: %v", err)
		return "fallback-jwt-secret-" + fmt.Sprintf("%d", time.Now().Unix())
	}

	// Base64 URL 인코딩으로 안전한 문자열 생성
	return base64.URLEncoding.EncodeToString(bytes)
}

// shouldIncludeExtraEnv는 추가로 포함할 환경변수인지 판단합니다.
func shouldIncludeExtraEnv(env string) bool {
	// 기본 설정에 이미 포함된 것들은 제외
	for key := range envDefaults {
		if strings.HasPrefix(env, key+"=") {
			return false
		}
	}

	// 시스템 환경변수는 제외
	systemVars := []string{"PATH=", "HOME=", "USER=", "HOSTNAME=", "PWD=", "SHLVL=", "_=", "TERM="}
	for _, sysVar := range systemVars {
		if strings.HasPrefix(env, sysVar) {
			return false
		}
	}

	// 애플리케이션 관련 환경변수만 포함
	appPrefixes := []string{"APP_", "CONFIG_", "CUSTOM_", "API_"}
	for _, prefix := range appPrefixes {
		if strings.HasPrefix(env, prefix) {
			return true
		}
	}

	return false
}

// validateConfig는 필수 설정값들이 있는지 검증합니다.
func validateConfig(cfg *Config) error {
	var missingFields []string

	if cfg.DBHost == "" {
		missingFields = append(missingFields, "DB_HOST")
	}
	if cfg.DBPort == "" {
		missingFields = append(missingFields, "DB_PORT")
	}
	if cfg.DBUser == "" {
		missingFields = append(missingFields, "DB_USER")
	}
	if cfg.DBName == "" {
		missingFields = append(missingFields, "DB_NAME")
	}
	if cfg.JWTSecret == "" {
		missingFields = append(missingFields, "JWT_SECRET")
	}

	// JWT 시크릿 길이 검증
	if cfg.JWTSecret != "" && len(cfg.JWTSecret) < 32 {
		log.Printf("보안 경고: JWT_SECRET이 너무 짧습니다 (최소 32자 권장). 현재 길이: %d", len(cfg.JWTSecret))
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("필수 설정값이 누락되었습니다: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// expandEnvVars는 문자열 내의 환경변수를 확장합니다 ($VAR 또는 ${VAR} 형식).
func expandEnvVars(value string) string {
	if value == "" {
		return value
	}

	return os.ExpandEnv(value)
}

// getEnv는 환경 변수를 읽고, 값이 없을 경우 기본값을 반환하는 헬퍼 함수입니다.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// LogConfig는 현재 설정을 로그로 출력합니다 (민감한 정보는 마스킹).
func (c *Config) LogConfig() {
	log.Println("=== 현재 애플리케이션 설정 ===")
	log.Printf("DB Host: %s", c.DBHost)
	log.Printf("DB Port: %s", c.DBPort)
	log.Printf("DB User: %s", c.DBUser)
	log.Printf("DB Name: %s", c.DBName)
	log.Printf("DB Password: %s", maskSensitive(c.DBPassword))
	log.Printf("JWT Secret: %s", maskSensitive(c.JWTSecret))
	log.Printf("Server Port: %s", c.ServerPort)
	log.Printf("Key Bits: %d", c.KeyBits)
	log.Printf("Auto Install Keys: %t", c.AutoInstallKeys)
	log.Printf("SSH User: %s", c.SSHUser)
	log.Printf("SSH Home Path: %s", c.SSHHomePath)
	log.Printf("Admin Username: %s", c.AdminUsername)
	log.Printf("Admin Password: %s", maskSensitive(c.AdminPassword))
	log.Println("===============================")
}

// maskSensitive는 민감한 정보를 마스킹합니다
func maskSensitive(value string) string {
	if value == "" {
		return "미설정"
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}
