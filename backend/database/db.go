package database

import (
	"time"

	"root/shared/logger"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/sharding"
)

const NumShards = 4 // Количество шардов

func ConnectDb(url string, log *logger.Logger) (*gorm.DB, error) {
	log.Debug("🔍 Starting database connection process")

	// Загружаем переменные окружения
	log.Debug("🔄 Loading environment variables from .env file")
	if err := godotenv.Load("../.env"); err != nil {
		log.Warnf("⚠️ Could not load .env file: %v. Using default environment variables.", err)
	}

	// Подключаемся к БД
	log.Debugf("🔗 Connecting to database with DSN: %s", url)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  url,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
	})
	if err != nil {
		log.Errorf("❌ Failed to connect to database: %v", err)
		return nil, err
	}

	log.Info("✅ Database connection established successfully")

	// Включаем шардирование
	log.Debug("⚡ Enabling database sharding")
	err = db.Use(sharding.Register(sharding.Config{
		ShardingKey:         "group_id",
		NumberOfShards:      NumShards,
		PrimaryKeyGenerator: sharding.PKSnowflake, // Генератор уникальных ID
	}, "songs"))
	if err != nil {
		log.Errorf("❌ Failed to register sharding: %v", err)
		return nil, err
	}

	log.Info("✅ Sharding configured successfully")

	// Настроим пул подключений
	log.Debug("📦 Configuring database connection pool")
	sqlDB, err := db.DB()
	if err != nil {
		log.Errorf("❌ Failed to configure database connection pool: %v", err)
		return nil, err
	}

	sqlDB.SetMaxIdleConns(20)
	log.Debug("⚙️ Max idle connections set to 20")

	sqlDB.SetMaxOpenConns(200)
	log.Debug("⚙️ Max open connections set to 200")

	sqlDB.SetConnMaxLifetime(time.Hour)
	log.Debugf("⚙️ Connection max lifetime set to %v", time.Hour)

	log.Info("✅ Database connection pool configured successfully")

	log.Debug("🔍 Database connection process completed")

	return db, nil
}
