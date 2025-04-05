package models

import "time"

type Phrase struct {
	ID                uint      `gorm:"primaryKey"`
	UserID            uint      `gorm:"not null;index"`
	TelegramMessageID int       `gorm:"not null;uniqueIndex"`
	Text              string    `gorm:"not null"`
	IsMastered        bool      `gorm:"not null;default:false"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	User              User      `gorm:"foreignKey:UserID"`
	Tags              []Tag     `gorm:"many2many:phrase_tag"`
}
