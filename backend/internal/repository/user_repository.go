// Package repository 封装数据访问。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
)

var ErrNotFound = errors.New("record not found")

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) List() ([]model.User, error) {
	var users []model.User
	err := r.db.Order("id DESC").Find(&users).Error
	return users, err
}

func (r *UserRepository) UpdateStatus(id int64, status string) error {
	return r.updateColumn(id, "status", status)
}

func (r *UserRepository) UpdateRole(id int64, role string) error {
	return r.updateColumn(id, "role", role)
}

func (r *UserRepository) UpdatePassword(id int64, hash string) error {
	return r.updateColumn(id, "password_hash", hash)
}

func (r *UserRepository) Delete(id int64) error {
	res := r.db.Delete(&model.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) updateColumn(id int64, column, value string) error {
	res := r.db.Model(&model.User{}).Where("id = ?", id).Update(column, value)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
