package song_repository

import (
	"context"
	"root/shared/logger"
	"gorm.io/gorm"
)

var _ ISongRepository = (*SongRepository)(nil)

type ISongRepository interface {
	CheckTable(ctx context.Context, groupName string) (int, error)
}

type SongRepository struct {
	logger *logger.Logger
	db     *gorm.DB
}

func NewSongRepository(logger *logger.Logger, db *gorm.DB) *SongRepository {
	return &SongRepository{
		logger: logger,
		db:     db,
	}
}
