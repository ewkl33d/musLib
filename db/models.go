package db

import (
	"gorm.io/gorm"
)

type Song struct {
	gorm.Model
	GroupName   string `gorm:"column:group_name"`
	SongName    string `gorm:"column:song_name"`
	ReleaseDate string `gorm:"column:release_date"`
	Text        string `gorm:"column:text"`
	Link        string `gorm:"column:link"`
}
