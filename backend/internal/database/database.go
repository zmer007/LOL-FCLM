// Package database 负责建立 GORM 连接并执行开发期自动迁移。
package database

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"roadside-league/backend/internal/model"
)

// Open 连接 MySQL，容器同时启动时数据库可能尚未就绪，做有限次重试。
func Open(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	for attempt := 1; attempt <= 15; attempt++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger:         logger.Default.LogMode(logger.Warn),
			TranslateError: true,
		})
		if err == nil {
			sqlDB, e := db.DB()
			if e == nil && sqlDB.Ping() == nil {
				break
			} else if e != nil {
				err = e
			} else {
				err = errPingFailed
			}
		}
		log.Printf("waiting for mysql (attempt %d/15): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.User{}, &model.Hero{}, &model.GameAsset{}, &model.Setting{},
		&model.RuneTree{}, &model.Rune{}, &model.BuildPlan{},
	); err != nil {
		return nil, err
	}
	return db, nil
}

var errPingFailed = &dbError{msg: "mysql ping failed"}

type dbError struct{ msg string }

func (e *dbError) Error() string { return e.msg }
