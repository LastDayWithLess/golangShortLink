package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Init(envFiles ...string) error {
	return godotenv.Load(envFiles...)
}

type DataBaseConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Password string
	Host     string
	Port     string
	DB       int
}

func LoadDBConfig() DataBaseConfig {
	return DataBaseConfig{
		User:     getEnv("DB_USER"),
		Password: getEnv("DB_PASSWORD"),
		Host:     getEnv("DB_HOST"),
		Port:     getEnv("DB_PORT"),
		DBName:   getEnv("DB_NAME"),
		SSLMode:  getEnv("DB_SSL_MODE"),
	}
}

func LoadRedisConfig() RedisConfig {
	db, err := strconv.Atoi(getEnv("REDIS_DB"))
	if err != nil {
		db = 0
	}

	return RedisConfig{
		Password: getEnv("REDIS_PASSWORD"),
		Host:     getEnv("REDIS_HOST"),
		Port:     getEnv("REDIS_PORT"),
		DB:       db,
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return ""
}
