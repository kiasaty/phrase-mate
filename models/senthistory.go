package models

import "time"

type SentHistory struct {
	ID       uint      `gorm:"primaryKey"`
	PhraseID uint      `gorm:"not null"`
	SentAt   time.Time `gorm:"autoCreateTime"`
}
