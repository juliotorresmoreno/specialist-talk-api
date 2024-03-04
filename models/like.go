package models

import (
	"time"

	"gorm.io/gorm"
)

type Like struct {
	ID         uint           `gorm:"primaryKey"`
	PostId     uint           `gorm:"not null"`
	Post       Post           `gorm:"foreignKey:PostId"`
	AuthorId   uint           `gorm:"not null"`
	Author     User           `gorm:"foreignKey:AuthorId"`
	CreationAt time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"type:timestamptz"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz"`
}

func (u Like) TableName() string {
	return "likes"
}
