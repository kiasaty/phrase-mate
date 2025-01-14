package database

import (
	"github.com/kiasaty/phrase-mate/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseClient interface {
	Migrate()

	CreateUser(*models.User) (*models.User, error)
	FindUserByTelegramID(telegramID int64) (*models.User, error)

	CreateTag(*models.Tag) (*models.Tag, error)
	FindTagByName(name string) (*models.Tag, error)

	CreatePhrase(*models.Phrase) (*models.Phrase, error)
	FindPhraseByMessageId(messageID int) (phrase *models.Phrase)
	UpdatePhrase(phrase *models.Phrase) error
	UpdatePhraseTags(*models.Phrase, *[]models.Tag) error
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

func (c *Client) Migrate() {
	c.DB.AutoMigrate(
		&models.User{},
		&models.Tag{},
		&models.Phrase{},
		&models.SentHistory{},
	)
}
