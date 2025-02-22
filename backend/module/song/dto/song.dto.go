package dto

import (
	"time"

	"github.com/google/uuid"
)

type Song struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	GroupID     int       `json:"groupid" gorm:"not null;index"`
	Song        string    `json:"song" gorm:"not null"`
	Text        string    `json:"text" gorm:"type:text"`
	Link        string    `json:"link" gorm:"type:text"`
	ReleaseDate time.Time `json:"releasedate" gorm:"type:date"`
}

type SongDetails struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type SongText struct {
	Text string `json:"text"`
}
