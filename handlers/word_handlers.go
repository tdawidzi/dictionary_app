package handlers

import (
	"fmt"

	"github.com/tdawidzi/dictionary_app/models"
	"github.com/tdawidzi/dictionary_app/utils"

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

// Adds Word to database
func AddWord(p graphql.ResolveParams) (interface{}, error) {
	word, _ := p.Args["word"].(string)
	language, _ := p.Args["language"].(string)

	newWord := models.Word{Word: word, Language: language}
	if err := utils.DB.Create(&newWord).Error; err != nil {
		return nil, fmt.Errorf("failed to add word: %w", err)
	}
	return newWord, nil
}

// Modify existing word in db
func UpdateWord(p graphql.ResolveParams) (interface{}, error) {
	oldWord, _ := p.Args["oldWord"].(string)
	language, _ := p.Args["language"].(string)
	newWord, _ := p.Args["newWord"].(string)

	// Check if word exists
	var word models.Word
	if err := utils.DB.Where("word = ? AND language = ?", oldWord, language).First(&word).Error; err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	// Modify and save word
	word.Word = newWord
	if err := utils.DB.Save(&word).Error; err != nil {
		return nil, fmt.Errorf("failed to update word: %w", err)
	}

	return word, nil
}

// Delete existing word from database
func DeleteWord(p graphql.ResolveParams) (interface{}, error) {
	wordValue, _ := p.Args["word"].(string)
	language, _ := p.Args["language"].(string)

	// Check if word exists
	var word models.Word
	if err := utils.DB.Where("word = ? AND language = ?", wordValue, language).First(&word).Error; err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	// Delete the word
	if err := utils.DB.Delete(&word).Error; err != nil {
		return nil, fmt.Errorf("failed to delete word: %w", err)
	}

	// Return true if succeeded
	return true, nil
}
