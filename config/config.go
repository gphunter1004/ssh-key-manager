package config

import (
	"encoding/json"
	"os"

	"github.com/joho/godotenv"
)

// Config는 애플리케이션의 모든 설정을 담는 구조체입니다.
type Config struct {
	DBHost     string `json:"-"`
	DBPort     string `json:"-"`
	DBUser     string `json:"-"`
	DBPassword string `json:"-"`
	DBName     string `json:"-"`
	JWTSecret  string `json:"-"`
	ServerPort string `json:"ServerPort"`
	KeyBits    int    `json:"KeyBits"`
}

// LoadConfig는 .env와 config.json 파일에서 설정을 로드합니다.
func LoadConfig() (*Config, error) {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
	}

	// config.json 파일 로드
	file, err := os.Open("config/config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
