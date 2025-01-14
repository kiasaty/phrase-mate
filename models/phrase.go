package models

import "time"

type Phrase struct {
	ID                uint      `gorm:"primaryKey"`
	UserID            uint      `gorm:"not null"`
	TelegramMessageID int       `gorm:"unique;not null"`
	Text              string    `gorm:"not null"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	User              User      `gorm:"foreignKey:UserID"`
	Tags              []Tag     `gorm:"many2many:phrase_tag"`
	History           []SentHistory
}
