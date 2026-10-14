package service

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
)

type AdminUserService struct {
	users *repository.UserRepository
}

func NewAdminUserService(users *repository.UserRepository) *AdminUserService {
	return &AdminUserService{users: users}
}

func (s *AdminUserService) List() ([]model.User, error) {
	return s.users.List()
}

// Add 管理员直接创建账号。
func (s *AdminUserService) Add(username, password, role string) (*model.User, error) {
	if n := utf8.RuneCountInString(username); n < 3 || n > 20 {
		return nil, &ValidationError{Msg: "用户名长度需为 3-20 个字符"}
	}
	if n := utf8.RuneCountInString(password); n < 6 || n > 32 {
		return nil, &ValidationError{Msg: "密码长度需为 6-32 个字符"}
	}
	if role != "user" && role != "admin" {
		return nil, &ValidationError{Msg: "角色不合法"}
	}
	if _, err := s.users.FindByUsername(username); err == nil {
		return nil, ErrUsernameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		Status:       model.UserStatusActive,
	}
	if err := s.users.Create(u); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return u, nil
}

func (s *AdminUserService) SetStatus(id int64, status string) error {
	if status != model.UserStatusActive && status != model.UserStatusDisabled {
		return &ValidationError{Msg: "状态不合法"}
	}
	return s.mapNotFound(s.users.UpdateStatus(id, status))
}

func (s *AdminUserService) SetRole(id int64, role string) error {
	if role != "user" && role != "admin" {
		return &ValidationError{Msg: "角色不合法"}
	}
	return s.mapNotFound(s.users.UpdateRole(id, role))
}

func (s *AdminUserService) ResetPassword(id int64, password string) error {
	if n := utf8.RuneCountInString(password); n < 6 || n > 32 {
		return &ValidationError{Msg: "密码长度需为 6-32 个字符"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.mapNotFound(s.users.UpdatePassword(id, string(hash)))
}

func (s *AdminUserService) Delete(id int64) error {
	return s.mapNotFound(s.users.Delete(id))
}

func (s *AdminUserService) mapNotFound(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrUserNotFound
	}
	return err
}
