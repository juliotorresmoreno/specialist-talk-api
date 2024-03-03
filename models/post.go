package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID         uint           `gorm:"primaryKey"`
	Content    string         `gorm:"type:varchar(1000);default:'';not null"`
	AuthorId   uint           `gorm:"not null"`
	Author     User           `gorm:"foreignKey:AuthorId"`
	CreationAt time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"type:timestamptz"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz"`
}
