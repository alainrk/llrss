package config

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	DBPath string
	Debug  bool
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		// TODO: Config this
		DBPath: "feeds.db",
		Debug:  true,
	}
}

func InitDatabase(config *DatabaseConfig) (*gorm.DB, error) {
	gormConfig := &gorm.Config{}
	if config.Debug {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(sqlite.Open(config.DBPath), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return db, nil
}
