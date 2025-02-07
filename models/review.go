package models

import "time"

type Review struct {
	PhraseID      uint          `gorm:"not null;idx_phrase_user;foreignKey"`
	UserID        uint          `gorm:"not null;idx_phrase_user;foreignKey"`
	SessionID     uint          `gorm:"not null;index"`
	RecallQuality RecallQuality `gorm:"not null"`
	EaseFactor    float64       `gorm:"not null"`
	Interval      uint16        `gorm:"not null"`
	ReviewedAt    time.Time     `gorm:"autoCreateTime"`
	NextReviewAt  time.Time     `gorm:""` // change to date, it should not be datetime
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
