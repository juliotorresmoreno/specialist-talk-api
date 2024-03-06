package models

import (
	"time"

	"gorm.io/gorm"
)

type Chat struct {
	ID         uint           `gorm:"primaryKey"`
	Name       string         `gorm:"not null"`
	OwnerId    uint           `gorm:"not null"`
	Owner      *User          `gorm:"foreignKey:OwnerId"`
	Chats      []ChatUser     `gorm:"foreignKey:ChatId"`
	CreationAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz"`
}

func (u Chat) TableName() string {
	return "chats"
}

type ChatUser struct {
	ID         uint           `gorm:"primaryKey"`
	ChatId     uint           `gorm:"not null"`
	Chat       Chat           `gorm:"foreignKey:ChatId"`
	UserId     uint           `gorm:"not null"`
	User       User           `gorm:"foreignKey:UserId"`
	CreationAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz"`
}

func (u ChatUser) TableName() string {
	return "chat_users"
}
