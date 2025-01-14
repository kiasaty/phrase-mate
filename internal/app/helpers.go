package app

import "strings"

func extractHashtags(text string) []string {
	words := strings.Fields(text)
	var hashtags []string
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			hashtags = append(hashtags, word)
		}
	}
	return hashtags
}
