package models

import (
	"gorm.io/gorm"
)

type Carousel struct {
	gorm.Model

	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	Rank     int    `json:"rank" gorm:"uniqueIndex;not null"`
	Link     string `json:"link"`
	Caption  string `json:"caption"`
}
