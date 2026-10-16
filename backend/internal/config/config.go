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
	AdminUsername  string // 启动时确保存在的管理员账号
	AdminPassword  string
	MaxImageMB     int // 图片上传大小上限（MB）
	MaxVideoMB     int // 投稿视频大小上限（MB）

	// SMTP 邮件发送（未配置则邮件功能不可用）
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string // 发件人显示地址，为空时使用 SMTPUser

	// 验证码
	CodeTTLSeconds    int // 验证码有效期（秒）
	CodeResendSeconds int // 同一邮箱重发间隔（秒）
}

func Load() Config {
	return Config{
		Addr:              getenv("APP_ADDR", ":9090"),
		DBDSN:             os.Getenv("DB_DSN"),
		ValkeyAddr:        getenv("VALKEY_ADDR", "valkey:6379"),
		MediaRoot:         getenv("MEDIA_ROOT", "/data/media"),
		JWTSecret:         getenv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpireHours:    getenvInt("JWT_EXPIRE_HOURS", 168),
		AdminUsername:     getenv("ADMIN_USERNAME", "admin"),
		AdminPassword:     getenv("ADMIN_PASSWORD", "admin12345"),
		MaxImageMB:        getenvInt("MAX_IMAGE_MB", 5),
		MaxVideoMB:        getenvInt("MAX_VIDEO_MB", 200),
		SMTPHost:          os.Getenv("SMTP_HOST"),
		SMTPPort:          getenvInt("SMTP_PORT", 587),
		SMTPUser:          os.Getenv("SMTP_USER"),
		SMTPPassword:      os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:          os.Getenv("SMTP_FROM"),
		CodeTTLSeconds:    getenvInt("CODE_TTL_SECONDS", 600),
		CodeResendSeconds: getenvInt("CODE_RESEND_SECONDS", 60),
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
