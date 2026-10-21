// Package config 读取服务运行所需的环境变量，全部带默认值。
package config

import (
	"fmt"
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
	AdminUsername  string // 启动时确保存在的管理员账号
	AdminPassword  string
	MaxImageMB     int // 图片上传大小上限（MB）
	MaxVideoMB     int // 投稿视频大小上限（MB）
}

func Load() Config {
	return Config{
		Addr:           getenv("APP_ADDR", ":9090"),
		DBDSN:          dbDSN(),
		ValkeyAddr:     getenv("VALKEY_ADDR", "valkey:6379"),
		MediaRoot:      getenv("MEDIA_ROOT", getenv("MEDIA_PATH", "/data/media")),
		JWTSecret:      getenv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpireHours: getenvInt("JWT_EXPIRE_HOURS", 168),
		AdminUsername:  getenv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getenv("ADMIN_PASSWORD", "admin12345"),
		MaxImageMB:     getenvInt("MAX_IMAGE_MB", 5),
		MaxVideoMB:     getenvInt("MAX_VIDEO_MB", 200),
	}
}

// dbDSN 返回数据库 DSN。优先使用 DB_DSN；否则由 DB_USER/DB_PASSWORD/DB_HOST/DB_PORT/DB_NAME 拼接。
func dbDSN() string {
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}
	user := getenv("DB_USER", "root")
	pass := os.Getenv("DB_PASSWORD")
	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "3306")
	name := getenv("DB_NAME", "lol")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, pass, host, port, name)
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
