package database

import (
	"fmt"
	song_model "root/module/song/dto"
	"root/shared/logger"

	"gorm.io/gorm"
)


func Migrate(db *gorm.DB, trigger bool, log *logger.Logger) error {
	if trigger {
		log.Info("📦 Starting database migration...")

		// Миграция основной таблицы Group
		log.Debug("🔍 Migrating Group model")
		if err := db.AutoMigrate(&song_model.Group{}); err != nil {
			log.Errorf("✖ Failed to migrate Group table: %v", err)
			return err
		}

		// Миграция шардированных таблиц Song
		log.Debug("🔍 Migrating Song model across shards")
		for i := 0; i < NumShards; i++ {
			tableName := fmt.Sprintf("songs_%d", i)
			if err := db.Table(tableName).AutoMigrate(&song_model.Song{}); err != nil {
				log.Errorf("✖ Failed to migrate shard table %s: %v", tableName, err)
				return err
			}
			log.Infof("✅ Shard table %s migrated successfully", tableName)
		}

		log.Info("✅ Database migration completed successfully")
	} else {
		log.Info("ℹ️ Migration trigger is disabled, skipping migration")
	}

	log.Info("✅ Database connection established successfully")
	return nil
}
