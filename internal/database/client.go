package database

import (
	"time"

	"github.com/kiasaty/phrase-mate/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseClient interface {
	Migrate()
	Transaction(fc func(tx DatabaseClient) error) error

	CreateUser(*models.User) (*models.User, error)
	FindUserByTelegramID(telegramID int64) (*models.User, error)

	CreateTag(*models.Tag) (*models.Tag, error)
	FindTagByName(name string) (*models.Tag, error)

	CreatePhrase(*models.Phrase) (*models.Phrase, error)
	FindPhraseByMessageId(messageID int) (phrase *models.Phrase)
	UpdatePhrase(phrase *models.Phrase) error
	UpdatePhraseTags(*models.Phrase, *[]models.Tag) error

	CreateSession(*models.Session) (*models.Session, error)
	UpdateSession(*models.Session) error
	FindActiveSession(userID uint) (*models.Session, error)

	CreateReview(review *models.Review) error
	CreateReviews(reviews []models.Review) error
	FindPhraseLastReview(userId, phraseId uint) (*models.Review, error)
	CountNotReviewedPhrasesBySessionId(sessionID uint) (int64, error)
	GetDueReviews(userID uint, now time.Time, limit int) ([]models.Review, error)
	GetNewReviews(userID uint, sessionID uint, limit int) ([]models.Review, error)
}

type Client struct {
	DB *gorm.DB
}

func NewDatabaseClient(databaseDSN string) (DatabaseClient, error) {
	db, err := gorm.Open(
		sqlite.Open(databaseDSN),
		&gorm.Config{},
	)

	if err != nil {
		return nil, err
	}

	client := &Client{
		DB: db,
	}

	return client, nil
}

func (c *Client) Transaction(fc func(tx DatabaseClient) error) error {
	return c.DB.Transaction(func(tx *gorm.DB) error {
		txClient := &Client{
			DB: tx,
		}
		return fc(txClient)
	})
}

func (c *Client) Migrate() {
	c.DB.AutoMigrate(
		&models.User{},
		&models.Tag{},
		&models.Phrase{},
		&models.Review{},
	)
}
