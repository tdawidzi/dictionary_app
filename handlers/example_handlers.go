package handlers

import (
	"errors"
	"fmt"

	"github.com/tdawidzi/dictionary_app/models"
	"github.com/tdawidzi/dictionary_app/utils"

	"github.com/graphql-go/graphql"
	"gorm.io/gorm"
)

// GetExamplesForWord fetches example sentences for a given word.
func GetExamplesForWord(p graphql.ResolveParams) (interface{}, error) {
	wordText, ok := p.Args["word"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid word input")
	}

	// Fetch the word by its text
	var word models.Word
	if err := utils.DB.Where("word = ?", wordText).First(&word).Error; err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	// Fetch all examples in one query
	var examples []models.Example
	if err := utils.DB.Where("word_id = ?", word.ID).Find(&examples).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch examples: %w", err)
	}

	return examples, nil
}

// AddExample adds an example to db associated with given word
func AddExample(p graphql.ResolveParams) (interface{}, error) {
	wordText, _ := p.Args["word"].(string)
	language, _ := p.Args["language"].(string)
	exampleText, _ := p.Args["example"].(string)

	var word models.Word
	if err := utils.DB.Where("word = ? AND language = ?", wordText, language).First(&word).Error; err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	// Check if record exists
	var existing models.Example
	if err := utils.DB.Where("example = ? AND word_id = ?", exampleText, word.ID).First(&existing).Error; err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query example: %w", err)
	}

	// Create new record
	example := models.Example{
		WordID:  word.ID,
		Example: exampleText,
	}
	if err := utils.DB.Create(&example).Error; err != nil {
		return nil, fmt.Errorf("failed to create example: %w", err)
	}
	return example, nil
}

// UpdateExample - modifies example text with given example id
func UpdateExample(p graphql.ResolveParams) (interface{}, error) {
	// Check for errors in id and convert it to integer
	id, ok := p.Args["id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid or missing ID")
	}

	// Check if argument example is given
	newExample, hasExample := p.Args["example"].(string)

	// Find example with given id
	var example models.Example
	if err := utils.DB.First(&example, id).Error; err != nil {
		return nil, fmt.Errorf("example not found: %w", err)
	}

	// Update example if new example is not nil
	if hasExample {
		example.Example = newExample
	}

	// Save changes
	if err := utils.DB.Save(&example).Error; err != nil {
		return nil, fmt.Errorf("failed to update example: %w", err)
	}

	return example, nil
}

// Deleting example with given id from db
func DeleteExample(p graphql.ResolveParams) (interface{}, error) {
	id, ok := p.Args["id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid or missing ID")
	}

	// Find example with given id
	var example models.Example
	if err := utils.DB.First(&example, id).Error; err != nil {
		return false, fmt.Errorf("example not found: %w", err)
	}

	if err := utils.DB.Delete(&example).Error; err != nil {
		return false, fmt.Errorf("failed to delete example: %w", err)
	}

	// Return true if succeeded
	return true, nil
}
