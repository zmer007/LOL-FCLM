// Package service 实现业务逻辑。
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
)

var (
	ErrUsernameTaken   = errors.New("username already taken")
	ErrEmailTaken      = errors.New("email already taken")
	ErrBadCredentials  = errors.New("invalid username or password")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserDisabled    = errors.New("user disabled")
	ErrInvalidCode     = errors.New("invalid or expired verification code")
	ErrCodeRateLimited = errors.New("verification code rate limited")
	ErrEmailRequired   = errors.New("email is required")
)

// ValidationError 携带可直接展示给用户的校验提示。
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

// Claims 是 JWT 中携带的业务字段。
type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"usr"`
	Role     string `json:"rol"`
	jwt.RegisteredClaims
}

type AuthService struct {
	users           *repository.UserRepository
	valkey          *redis.Client
	mail            *MailService
	jwtSecret       []byte
	expireHours     int
	codeTTL         time.Duration
	codeResendGap   time.Duration
}

func NewAuthService(users *repository.UserRepository, valkey *redis.Client, mail *MailService, jwtSecret string, expireHours, codeTTLSeconds, codeResendSeconds int) *AuthService {
	return &AuthService{
		users:         users,
		valkey:        valkey,
		mail:          mail,
		jwtSecret:     []byte(jwtSecret),
		expireHours:   expireHours,
		codeTTL:       time.Duration(codeTTLSeconds) * time.Second,
		codeResendGap: time.Duration(codeResendSeconds) * time.Second,
	}
}

// SendVerificationCode 生成 6 位验证码，存 Valkey 并通过邮件发送。
func (s *AuthService) SendVerificationCode(email string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	if !s.mail.Configured() {
		return ErrMailNotConfigured
	}

	ctx := context.Background()
	rateKey := "code:rate:" + email
	// 频率限制：60s 内只能发一次（SetNX 原子操作避免并发穿透）
	ok, err := s.valkey.SetNX(ctx, rateKey, "1", s.codeResendGap).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrCodeRateLimited
	}

	code, err := generateCode()
	if err != nil {
		_ = s.valkey.Del(ctx, rateKey).Err()
		return err
	}

	verifyKey := "code:verify:" + email
	// 先存验证码，再发邮件；邮件失败时回删验证码
	if err := s.valkey.Set(ctx, verifyKey, code, s.codeTTL).Err(); err != nil {
		_ = s.valkey.Del(ctx, rateKey).Err()
		return err
	}
	if err := s.mail.SendVerificationCode(email, code); err != nil {
		_ = s.valkey.Del(ctx, verifyKey).Err()
		_ = s.valkey.Del(ctx, rateKey).Err()
		return err
	}
	return nil
}

// Register 校验邮箱验证码、创建用户并返回用户与访问令牌。
func (s *AuthService) Register(username, email, password, code string) (*model.User, string, error) {
	if err := validateUsername(username); err != nil {
		return nil, "", err
	}
	if err := validateEmail(email); err != nil {
		return nil, "", err
	}
	if err := validatePassword(password); err != nil {
		return nil, "", err
	}
	if code == "" {
		return nil, "", ErrInvalidCode
	}

	// 校验验证码
	ctx := context.Background()
	verifyKey := "code:verify:" + email
	stored, err := s.valkey.Get(ctx, verifyKey).Result()
	if err != nil {
		return nil, "", ErrInvalidCode
	}
	if stored != code {
		return nil, "", ErrInvalidCode
	}
	// 一次性：验证通过后删除
	_ = s.valkey.Del(ctx, verifyKey).Err()

	// 用户名 / 邮箱唯一性
	if _, err := s.users.FindByUsername(username); err == nil {
		return nil, "", ErrUsernameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, "", err
	}
	if _, err := s.users.FindByEmail(email); err == nil {
		return nil, "", ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	u := &model.User{
		Username:      username,
		Email:         email,
		EmailVerified: true,
		PasswordHash:  string(hash),
		Role:          "user",
		Status:        model.UserStatusActive,
	}
	if err := s.users.Create(u); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			// 并发场景下唯一键冲突，判断是用户名还是邮箱
			if _, e := s.users.FindByUsername(username); e == nil {
				return nil, "", ErrUsernameTaken
			}
			if _, e := s.users.FindByEmail(email); e == nil {
				return nil, "", ErrEmailTaken
			}
			return nil, "", ErrUsernameTaken
		}
		return nil, "", err
	}

	token, err := s.issueToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

// Login 校验账密并返回用户与访问令牌。account 可为用户名或邮箱。
func (s *AuthService) Login(account, password string) (*model.User, string, error) {
	if account == "" || password == "" {
		return nil, "", &ValidationError{Msg: "账号和密码不能为空"}
	}
	u, err := s.users.FindByUsername(account)
	if errors.Is(err, repository.ErrNotFound) {
		// 再按邮箱查找
		u, err = s.users.FindByEmail(account)
	}
	if errors.Is(err, repository.ErrNotFound) {
		return nil, "", ErrBadCredentials
	}
	if err != nil {
		return nil, "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, "", ErrBadCredentials
	}
	if u.Status == model.UserStatusDisabled {
		return nil, "", ErrUserDisabled
	}
	// 防御性约束：邮箱未验证的账号不可登录（admin 等无邮箱账号除外）
	if u.Email != "" && !u.EmailVerified {
		return nil, "", &ValidationError{Msg: "邮箱未验证，请先完成邮箱验证"}
	}
	token, err := s.issueToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

// EnsureAdmin 启动时确保配置中的管理员账号存在。
func (s *AuthService) EnsureAdmin(username, password string) error {
	if username == "" || password == "" {
		return nil
	}
	if _, err := s.users.FindByUsername(username); err == nil {
		return nil // 已存在
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.Create(&model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "admin",
		Status:       model.UserStatusActive,
	})
}

// IsActive 判断用户当前是否存在且未被禁用（供鉴权中间件实时校验）。
func (s *AuthService) IsActive(id int64) bool {
	u, err := s.users.FindByID(id)
	return err == nil && u.Status == model.UserStatusActive
}

// GetUser 按 ID 取用户（/me 使用）。
func (s *AuthService) GetUser(id int64) (*model.User, error) {
	u, err := s.users.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	return u, err
}

// ParseToken 校验并解析访问令牌。
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}

func (s *AuthService) issueToken(u *model.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   u.ID,
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

// generateCode 生成 6 位数字验证码（加密随机源）。
func generateCode() (string, error) {
	const digits = "0123456789"
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b[i] = digits[n.Int64()]
	}
	return string(b), nil
}

func validateUsername(username string) error {
	n := utf8.RuneCountInString(username)
	if n < 3 || n > 20 {
		return &ValidationError{Msg: "用户名长度需为 3-20 个字符"}
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return &ValidationError{Msg: "邮箱不能为空"}
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return &ValidationError{Msg: "邮箱格式不正确"}
	}
	return nil
}

func validatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if n < 6 || n > 32 {
		return &ValidationError{Msg: "密码长度需为 6-32 个字符"}
	}
	return nil
}
