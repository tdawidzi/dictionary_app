package handlers

import (
	"dictionary-app/models"
	"dictionary-app/utils"
	"fmt"
	"sync"

	"github.com/graphql-go/graphql"
)

// GetTranslationsForWord returns a list of words in a different language based on the given word.
func GetTranslationsForWord(p graphql.ResolveParams) (interface{}, error) {
	wordText, ok := p.Args["word"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid word")
	}

	// Retrieve the word from the database based on its value
	var word models.Word
	err := utils.DB.Where("word = ?", wordText).First(&word).Error
	if err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	var translations []models.Translation
	if word.Language == "pl" {
		// If the word is Polish, fetch English translations
		err = utils.DB.Where("word_id_pl = ?", word.ID).Find(&translations).Error
	} else if word.Language == "en" {
		// If the word is English, fetch Polish translations
		err = utils.DB.Where("word_id_en = ?", word.ID).Find(&translations).Error
	} else {
		return nil, fmt.Errorf("word has an unsupported language: %s", word.Language)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch translations: %w", err)
	}

	// Using concurrency to fetch the translated words
	var wg sync.WaitGroup
	resultCh := make(chan models.Word, len(translations))
	errCh := make(chan error, len(translations))

	for _, translation := range translations {
		wg.Add(1)
		go func(t models.Translation) {
			defer wg.Done()
			var translatedWord models.Word
			var lookupErr error

			if word.Language == "pl" {
				lookupErr = utils.DB.First(&translatedWord, t.WordIDEn).Error
			} else {
				lookupErr = utils.DB.First(&translatedWord, t.WordIDPl).Error
			}

			if lookupErr != nil {
				errCh <- lookupErr
			} else {
				resultCh <- translatedWord
			}
		}(translation)
	}

	// Closing channels once all goroutines finish
	go func() {
		wg.Wait()
		close(resultCh)
		close(errCh)
	}()

	var translatedWords []models.Word
	for translatedWord := range resultCh {
		translatedWords = append(translatedWords, translatedWord)
	}

	// Check if any errors occurred
	if len(errCh) > 0 {
		return nil, fmt.Errorf("one or more translations failed")
	}

	return translatedWords, nil
}
