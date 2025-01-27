package dto

type Song struct {
	ID          uint   `gorm:"primaryKey"`
	Group       string `gorm:"type:varchar(255);not null"`
	Song        string `gorm:"type:varchar(255);not null"`
	ReleaseDate string `gorm:"type:varchar(255)"`
	Text        string `gorm:"type:text"`
	Link        string `gorm:"type:varchar(255)"`
}

type SongDetails struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type SongText struct{
	Text string `json:"text"`
}