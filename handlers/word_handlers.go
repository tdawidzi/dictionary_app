package handlers

import (
	"dictionary-app/models"
	"dictionary-app/utils"
	"fmt"

	"github.com/graphql-go/graphql"
)

// GetWords fetches all words in database - for display of dictionary content
func GetWords(p graphql.ResolveParams) (interface{}, error) {
	var words []models.Word
	err := utils.DB.Find(&words).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch words: %w", err)
	}
	return words, nil
}
