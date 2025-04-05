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

	var word models.Word
	if err := utils.DB.Where("word = ?", wordText).First(&word).Error; err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	// Early return for unsupported languages
	if word.Language != "pl" && word.Language != "en" {
		return nil, fmt.Errorf("unsupported language: %s", word.Language)
	}

	// Fetch translations based on word language
	var translations []models.Translation
	var translationErr error

	switch word.Language {
	case "pl":
		translationErr = utils.DB.Where("word_id_pl = ?", word.ID).Find(&translations).Error
	case "en":
		translationErr = utils.DB.Where("word_id_en = ?", word.ID).Find(&translations).Error
	default:
		return nil, fmt.Errorf("unsupported language: %s", word.Language)
	}

	if translationErr != nil {
		return nil, fmt.Errorf("failed to fetch translations: %w", translationErr)
	}

	// Pre-allocate slice with known capacity
	translatedWords := make([]models.Word, 0, len(translations))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errorsOccurred bool

	for _, translation := range translations {
		wg.Add(1)
		go func(t models.Translation) {
			defer wg.Done()

			var translatedWord models.Word
			var lookupErr error

			// Determine which word ID to look up
			wordID := t.WordIDEn
			if word.Language == "en" {
				wordID = t.WordIDPl
			}

			lookupErr = utils.DB.First(&translatedWord, wordID).Error

			mu.Lock()
			defer mu.Unlock()

			if lookupErr != nil {
				errorsOccurred = true
			} else {
				translatedWords = append(translatedWords, translatedWord)
			}
		}(translation)
	}

	wg.Wait()

	if errorsOccurred {
		return nil, fmt.Errorf("some translations could not be fetched")
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
