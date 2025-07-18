package config

import (
	"log"
	"os"
	"strconv"

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
}

// LoadConfig는 .env 파일에서 모든 설정을 로드합니다.
func LoadConfig() (*Config, error) {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		log.Printf("경고: .env 파일을 찾을 수 없습니다. 환경 변수를 직접 사용합니다.")
	}

	cfg := &Config{
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		ServerPort:  getEnv("SERVER_PORT", "8080"), // 기본값 설정
		SSHUser:     os.Getenv("SSH_USER"),
		SSHHomePath: os.Getenv("SSH_HOME_PATH"),
	}

	// KeyBits 파싱 (정수형)
	keyBitsStr := getEnv("KEY_BITS", "4096")
	keyBits, err := strconv.Atoi(keyBitsStr)
	if err != nil {
		log.Printf("경고: KEY_BITS 파싱 실패, 기본값(4096) 사용. 오류: %v", err)
		cfg.KeyBits = 4096 // 파싱 실패 시 기본값
	} else {
		cfg.KeyBits = keyBits
	}

	// AutoInstallKeys 파싱 (불리언)
	autoInstallStr := getEnv("AUTO_INSTALL_KEYS", "false")
	autoInstall, err := strconv.ParseBool(autoInstallStr)
	if err != nil {
		log.Printf("경고: AUTO_INSTALL_KEYS 파싱 실패, 기본값(false) 사용. 오류: %v", err)
		cfg.AutoInstallKeys = false // 파싱 실패 시 기본값
	} else {
		cfg.AutoInstallKeys = autoInstall
	}

	return cfg, nil
}

// getEnv는 환경 변수를 읽고, 값이 없을 경우 기본값을 반환하는 헬퍼 함수입니다.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
