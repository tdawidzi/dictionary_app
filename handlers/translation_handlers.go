package handlers

import (
	"fmt"
	"sync"

	"github.com/tdawidzi/dictionary_app/models"
	"github.com/tdawidzi/dictionary_app/utils"

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

	// Using concurrency to fetch the translated words simultaneously
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

	// Check for errors
	if len(errCh) > 0 {
		return nil, fmt.Errorf("one or more translations failed")
	}

	return translatedWords, nil
}

// Adds translation to db
func AddTranslation(p graphql.ResolveParams) (interface{}, error) {
	wordPl, _ := p.Args["wordPl"].(string)
	wordEn, _ := p.Args["wordEn"].(string)

	var wordPL models.Word
	var wordEN models.Word

	// Check if words exists in dictionary with given languages
	if err := utils.DB.Where("word = ? AND language = 'pl'", wordPl).First(&wordPL).Error; err != nil {
		return nil, fmt.Errorf("polish word not found: %w", err)
	}

	if err := utils.DB.Where("word = ? AND language = 'en'", wordEn).First(&wordEN).Error; err != nil {
		return nil, fmt.Errorf("english word not found: %w", err)
	}

	// Get word id's
	translation := models.Translation{
		WordIDPl: wordPL.ID,
		WordIDEn: wordEN.ID,
	}

	// Add translation
	if err := utils.DB.Create(&translation).Error; err != nil {
		return nil, fmt.Errorf("failed to create translation: %w", err)
	}

	return translation, nil
}

// Modify translation existing in db
func UpdateTranslation(p graphql.ResolveParams) (interface{}, error) {
	oldWordPl, _ := p.Args["oldWordPl"].(string)
	oldWordEn, _ := p.Args["oldWordEn"].(string)
	newWordPl, _ := p.Args["newWordPl"].(string)
	newWordEn, _ := p.Args["newWordEn"].(string)

	var oldPL, oldEN, newPL, newEN models.Word

	// Check if all given words exists in db
	if err := utils.DB.Where("word = ? AND language = 'pl'", oldWordPl).First(&oldPL).Error; err != nil {
		return nil, fmt.Errorf("old Polish word not found: %w", err)
	}

	if err := utils.DB.Where("word = ? AND language = 'en'", oldWordEn).First(&oldEN).Error; err != nil {
		return nil, fmt.Errorf("old English word not found: %w", err)
	}

	if err := utils.DB.Where("word = ? AND language = 'pl'", newWordPl).First(&newPL).Error; err != nil {
		return nil, fmt.Errorf("new Polish word not found: %w", err)
	}

	if err := utils.DB.Where("word = ? AND language = 'en'", newWordEn).First(&newEN).Error; err != nil {
		return nil, fmt.Errorf("new English word not found: %w", err)
	}

	// Check if old translation exists
	var translation models.Translation
	if err := utils.DB.Where("word_id_pl = ? AND word_id_en = ?", oldPL.ID, oldEN.ID).First(&translation).Error; err != nil {
		return nil, fmt.Errorf("translation not found: %w", err)
	}

	// Modify and save translation
	translation.WordIDPl = newPL.ID
	translation.WordIDEn = newEN.ID

	if err := utils.DB.Model(&models.Translation{}).
		Where("id = ?", translation.ID).
		Updates(models.Translation{
			WordIDPl: newPL.ID,
			WordIDEn: newEN.ID,
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to update translation: %w", err)
	}

	return translation, nil
}

// Delete existing translation from db
func DeleteTranslation(p graphql.ResolveParams) (interface{}, error) {
	wordPl, _ := p.Args["wordPl"].(string)
	wordEn, _ := p.Args["wordEn"].(string)

	var wordPL models.Word
	var wordEN models.Word

	// Check for words in db
	if err := utils.DB.Where("word = ? AND language = 'pl'", wordPl).First(&wordPL).Error; err != nil {
		return nil, fmt.Errorf("polish word not found: %w", err)
	}

	if err := utils.DB.Where("word = ? AND language = 'en'", wordEn).First(&wordEN).Error; err != nil {
		return nil, fmt.Errorf("english word not found: %w", err)
	}

	// Delete translation
	if err := utils.DB.Where("word_id_pl = ? AND word_id_en = ?", wordPL.ID, wordEN.ID).Delete(&models.Translation{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete translation: %w", err)
	}

	return true, nil
}
