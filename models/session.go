package models

import "time"

type Session struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    uint       `gorm:"not null"`
	StartedAt time.Time  `gorm:"autoCreateTime"`
	EndedAt   *time.Time `gorm:""`
}
