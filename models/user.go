package models

import "time"

type User struct {
	ID             uint      `gorm:"primaryKey"`
	TelegramChatID int64     `gorm:"unique;not null"`
	FirstName      string    `gorm:"size:100"`
	LastName       string    `gorm:"size:100"`
	Username       string    `gorm:"size:100"`
	LanguageCode   string    `gorm:"size:10"`
	IsBot          bool      `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}
