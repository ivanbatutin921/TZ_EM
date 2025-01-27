package database

import (
	song_model "root/module/song/dto"
	"root/shared/logger"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, trigger bool, log *logger.Logger) error {
	if trigger {
		log.Info("📦 Starting database migration...")
		models := []interface{}{
			&song_model.Song{},
		}

		log.Debug("🔍 Models to migrate: song_model.Song")
		if err := db.AutoMigrate(models...); err != nil {
			log.Errorf("✖ Failed to migrate database: %v", err)
			return err
		}
		log.Info("✅ Database migration completed successfully")
	} else {
		log.Info("ℹ️ Migration trigger is disabled, skipping migration")
	}

	log.Info("✅ Database connection established successfully")
	return nil
}
