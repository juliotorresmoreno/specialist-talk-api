package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID         uint           `gorm:"primaryKey"`
	Content    string         `gorm:"not null"`
	PostId     uint           `gorm:"not null"`
	Post       Post           `gorm:"foreignKey:PostId"`
	AuthorId   uint           `gorm:"not null"`
	Author     User           `gorm:"foreignKey:AuthorId"`
	CreationAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz"`
}

func (u Comment) TableName() string {
	return "comments"
}
