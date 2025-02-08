package models

import "time"

type Review struct {
	ID            uint          `gorm:"primaryKey"`
	PhraseID      uint          `gorm:"not null;idx_phrase_user;foreignKey"`
	UserID        uint          `gorm:"not null;idx_phrase_user;foreignKey"`
	SessionID     uint          `gorm:"not null;index"`
	RecallQuality RecallQuality `gorm:"not null"`
	EaseFactor    float64       `gorm:"not null"`
	Interval      uint16        `gorm:"not null"`
	ReviewedAt    *time.Time    `gorm:"type:datetime"`
	NextReviewAt  *time.Time    `gorm:"type:date"`
}

type RecallQuality uint8

const (
	QualityForgot     RecallQuality = 1
	QualityHesitant   RecallQuality = 2
	QualityRemembered RecallQuality = 3
	QualityFluent     RecallQuality = 4
	QualityPerfect    RecallQuality = 5
)

func (q RecallQuality) IsValid() bool {
	return q >= QualityForgot && q <= QualityPerfect
}
