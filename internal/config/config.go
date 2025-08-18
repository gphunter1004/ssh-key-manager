package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config는 애플리케이션 설정 구조체입니다.
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Server
	JWTSecret  string
	ServerPort string
	KeyBits    int

	// SSH 설정
	AutoInstallKeys bool
	SSHUser         string
	SSHHomePath     string

	// Admin
	AdminUsername string
	AdminPassword string
}

// Load는 설정을 로드합니다.
func Load() (*Config, error) {
	// .env 파일 로드 (없어도 에러 아님)
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:        getEnv("DB_HOST", "postgres"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "password"),
		DBName:        getEnv("DB_NAME", "key-manager"),
		JWTSecret:     getEnv("JWT_SECRET", generateSecret()),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		SSHUser:       getEnv("SSH_USER", "robos"),
		SSHHomePath:   os.ExpandEnv(getEnv("SSH_HOME_PATH", "/home/$SSH_USER")),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}

	// KeyBits 파싱
	if keyBitsStr := getEnv("KEY_BITS", "4096"); keyBitsStr != "" {
		if keyBits, err := strconv.Atoi(keyBitsStr); err == nil {
			cfg.KeyBits = keyBits
		} else {
			cfg.KeyBits = 4096
		}
	}

	// AutoInstallKeys 파싱
	cfg.AutoInstallKeys = strings.ToLower(getEnv("AUTO_INSTALL_KEYS", "false")) == "true"

	return cfg, nil
}

// GetDSN은 데이터베이스 연결 문자열을 반환합니다.
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func generateSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback-secret-" + fmt.Sprintf("%d", os.Getpid())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
