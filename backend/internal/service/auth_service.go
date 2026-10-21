// Package service 实现业务逻辑。
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
)

var (
	ErrUsernameTaken  = errors.New("username already taken")
	ErrBadCredentials = errors.New("invalid username or password")
	ErrUserNotFound   = errors.New("user not found")
	ErrUserDisabled   = errors.New("user disabled")
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
	users       *repository.UserRepository
	valkey      *redis.Client
	jwtSecret   []byte
	expireHours int
}

func NewAuthService(users *repository.UserRepository, valkey *redis.Client, jwtSecret string, expireHours int) *AuthService {
	return &AuthService{
		users:       users,
		valkey:      valkey,
		jwtSecret:   []byte(jwtSecret),
		expireHours: expireHours,
	}
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
