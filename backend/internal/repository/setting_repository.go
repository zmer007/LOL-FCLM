package repository

import (
	"errors"

	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) Get(key string) (string, error) {
	var s model.Setting
	err := r.db.First(&s, "`key` = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return s.Value, nil
}

func (r *SettingRepository) Set(key, value string) error {
	return r.db.Save(&model.Setting{Key: key, Value: value}).Error
}
