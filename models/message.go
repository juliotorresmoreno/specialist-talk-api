package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID         uint           `gorm:"primaryKey"`
	ChatId     uint           `gorm:"not null"`
	Chat       Chat           `gorm:"foreignKey:ChatId"`
	UserId     uint           `gorm:"not null"`
	User       User           `gorm:"foreignKey:UserId"`
	Content    string         `gorm:"not null"`
	CreationAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz"`
}
