package config

import (
	"os"
	"strconv"
	"time"
)

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func GetEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

type Base struct {
	Port        string
	Environment string
	LogLevel    string
}

func LoadBase(defaultPort string) Base {
	return Base{
		Port:        GetEnv("PORT", defaultPort),
		Environment: GetEnv("ENVIRONMENT", "development"),
		LogLevel:    GetEnv("LOG_LEVEL", "info"),
	}
}

type Postgres struct {
	URL string
}

func LoadPostgres() Postgres {
	return Postgres{URL: GetEnv("DATABASE_URL", "postgres://markettg:markettg@postgres:5432/markettg?sslmode=disable")}
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}

func LoadRedis() Redis {
	return Redis{
		Addr:     GetEnv("REDIS_ADDR", "redis:6379"),
		Password: GetEnv("REDIS_PASSWORD", ""),
		DB:       GetEnvInt("REDIS_DB", 0),
	}
}

type S3 struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	PublicURL string
	UseSSL    bool
}

func LoadS3() S3 {
	return S3{
		Endpoint:  GetEnv("S3_ENDPOINT", "minio:9000"),
		AccessKey: GetEnv("S3_ACCESS_KEY", "minioadmin"),
		SecretKey: GetEnv("S3_SECRET_KEY", "minioadmin"),
		Bucket:    GetEnv("S3_BUCKET", "catalog"),
		PublicURL: GetEnv("S3_PUBLIC_URL", "http://localhost:9000/catalog"),
		UseSSL:    GetEnv("S3_USE_SSL", "false") == "true",
	}
}
