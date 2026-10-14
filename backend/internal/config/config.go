// Package config 读取服务运行所需的环境变量，全部带默认值。
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr           string // API 监听地址
	DBDSN          string // MySQL DSN
	ValkeyAddr     string // Valkey/Redis 地址
	MediaRoot      string // 上传文件目录
	JWTSecret      string // JWT 签名密钥
	JWTExpireHours int    // JWT 有效期（小时）
}

func Load() Config {
	return Config{
		Addr:           getenv("APP_ADDR", ":9090"),
		DBDSN:          os.Getenv("DB_DSN"),
		ValkeyAddr:     getenv("VALKEY_ADDR", "valkey:6379"),
		MediaRoot:      getenv("MEDIA_ROOT", "/data/media"),
		JWTSecret:      getenv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpireHours: getenvInt("JWT_EXPIRE_HOURS", 168),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
