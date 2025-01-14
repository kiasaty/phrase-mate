package database

import (
	"strings"

	"github.com/kiasaty/phrase-mate/models"
)

func (c *Client) CreateTag(tag *models.Tag) (*models.Tag, error) {
	tag.Name = strings.ToLower(tag.Name)

	if err := c.DB.Create(tag).Error; err != nil {
		return nil, err
	}

	return tag, nil
}

func (c *Client) FindTagByName(name string) (*models.Tag, error) {
	var tag models.Tag

	if err := c.DB.Where("LOWER(name) = ?", strings.ToLower(name)).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}
