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

	CreateUser(user *models.User) (*models.User, error)
	FindUserByTelegramID(telegramID int64) (*models.User, error)
	GetAllUsers() ([]*models.User, error)

	CreateTag(tag *models.Tag) (*models.Tag, error)
	FindTagByName(name string) (*models.Tag, error)

	CreatePhrase(phrase *models.Phrase) (*models.Phrase, error)
	FindPhrase(phraseID uint) (*models.Phrase, error)
	FindPhraseByMessageId(messageID int) (phrase *models.Phrase)
	UpdatePhrase(phrase *models.Phrase) error
	UpdatePhraseTags(phrase *models.Phrase, tags *[]models.Tag) error
	FindNewPhrasesToReview(userID uint, limit int) ([]uint, error)

	CreateSession(session *models.Session) (*models.Session, error)
	EndSession(sessionID uint) error
	FindActiveSession(userID uint) (*models.Session, error)

	CreateReview(review *models.Review) error
	UpdateReview(review *models.Review) error
	FindReview(userID uint, phraseId uint) (*models.Review, error)
	CountReviewedPhrasesInSession(sessionID uint) (uint, error)
	GetDueReview(userID uint, now time.Time, limit uint) (*models.Review, error)

	// Review history operations
	CreateReviewHistory(review *models.ReviewHistory) error
	FindReviewHistory(userID uint, phraseID uint) ([]*models.ReviewHistory, error)
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
		&models.ReviewHistory{},
		&models.Session{},
	)
}
