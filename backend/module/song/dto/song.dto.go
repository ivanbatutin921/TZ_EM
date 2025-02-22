package dto

import (
	"time"

	"github.com/google/uuid"
)

type Song struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	GroupID     int       `json:"groupid" gorm:"not null;index"`
	Group       string    `json:"group" gorm:"not null;index"`
	Song        string    `json:"song" gorm:"not null"`
	Text        string    `json:"text" gorm:"type:text;index"`
	Link        string    `json:"link" gorm:"type:text;index"`
	ReleaseDate time.Time `json:"releasedate" gorm:"type:date;index"`
}

type SongDetails struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type SongText struct {
	Text string `json:"text"`
}
