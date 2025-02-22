package dto

type Group struct {
	ID   int   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}