package handlers

import (
	"errors"
	"fmt"

	"github.com/tdawidzi/dictionary_app/models"
	"github.com/tdawidzi/dictionary_app/utils"

	"github.com/graphql-go/graphql"
	"gorm.io/gorm"
)

func GetTranslationsForWord(p graphql.ResolveParams) (interface{}, error) {
	word, ok := p.Source.(models.Word)
	if !ok {
		return nil, fmt.Errorf("invalid source for translations")
	}

	var translations []models.Translation
	var err error

	if word.Language == "pl" {
		err = utils.DB.Preload("WordEn").
			Where("word_id_pl = ?", word.ID).
			Find(&translations).Error
	} else if word.Language == "en" {
		err = utils.DB.Preload("WordPl").
			Where("word_id_en = ?", word.ID).
			Find(&translations).Error
	} else {
		return nil, fmt.Errorf("unsupported language: %s", word.Language)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch translations: %w", err)
	}

	translatedWords := make([]models.Word, 0, len(translations))
	for _, t := range translations {
		if word.Language == "pl" {
			translatedWords = append(translatedWords, t.WordEn)
		} else {
			translatedWords = append(translatedWords, t.WordPl)
		}
	}

	return translatedWords, nil
}

// Adds translation to db
func AddTranslation(p graphql.ResolveParams) (interface{}, error) {
	wordPl, _ := p.Args["wordPl"].(string)
	wordEn, _ := p.Args["wordEn"].(string)

	var wordPL models.Word
	var wordEN models.Word

	// Check if words exists
	if err := utils.DB.Where("word = ? AND language = 'pl'", wordPl).First(&wordPL).Error; err != nil {
		return nil, fmt.Errorf("polish word not found: %w", err)
	}
	if err := utils.DB.Where("word = ? AND language = 'en'", wordEn).First(&wordEN).Error; err != nil {
		return nil, fmt.Errorf("english word not found: %w", err)
	}

	// Check if translation exists
	var existing models.Translation
	if err := utils.DB.Where("word_id_pl = ? AND word_id_en = ?", wordPL.ID, wordEN.ID).First(&existing).Error; err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query translation: %w", err)
	}

	// Create new translation
	translation := models.Translation{
		WordIDPl: wordPL.ID,
		WordIDEn: wordEN.ID,
	}
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

	// Check if new translation does not exist
	var existing models.Translation
	if err := utils.DB.Where("word_id_pl = ? AND word_id_en = ?", newPL.ID, newEN.ID).First(&existing).Error; err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query translation: %w", err)
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
