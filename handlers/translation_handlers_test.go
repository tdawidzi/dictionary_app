package handlers_test

import (
	"sync"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/tdawidzi/dictionary_app/handlers"
	"github.com/tdawidzi/dictionary_app/models"
	"github.com/tdawidzi/dictionary_app/testresources"
	"github.com/tdawidzi/dictionary_app/utils"
)

func setupTranslationTestDB(t *testing.T) {
	utils.DB = testresources.NewSingleTestConnection(t)
	err := utils.DB.AutoMigrate(&models.Word{}, &models.Translation{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}

func TestAddAndGetTranslations(t *testing.T) {
	setupTranslationTestDB(t)

	pl := models.Word{Word: "kot", Language: "pl"}
	en := models.Word{Word: "cat", Language: "en"}
	utils.DB.Create(&pl)
	utils.DB.Create(&en)

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"wordPl": "kot",
			"wordEn": "cat",
		},
	}

	// Add translation
	result, err := handlers.AddTranslation(params)
	assert.NoError(t, err)
	translation, ok := result.(models.Translation)
	assert.True(t, ok)
	assert.Equal(t, pl.ID, translation.WordIDPl)
	assert.Equal(t, en.ID, translation.WordIDEn)

	// Get translations for "kot"
	sourceParams := graphql.ResolveParams{Source: pl}
	result, err = handlers.GetTranslationsForWord(sourceParams)
	assert.NoError(t, err)
	translations, ok := result.([]models.Word)
	assert.True(t, ok)
	assert.Len(t, translations, 1)
	assert.Equal(t, "cat", translations[0].Word)
}

func TestUpdateTranslation(t *testing.T) {
	setupTranslationTestDB(t)

	oldPl := models.Word{Word: "pies", Language: "pl"}
	oldEn := models.Word{Word: "dog", Language: "en"}
	newPl := models.Word{Word: "kundel", Language: "pl"}
	newEn := models.Word{Word: "mongrel", Language: "en"}

	utils.DB.Create(&oldPl)
	utils.DB.Create(&oldEn)
	utils.DB.Create(&newPl)
	utils.DB.Create(&newEn)

	// Create old translation
	utils.DB.Create(&models.Translation{WordIDPl: oldPl.ID, WordIDEn: oldEn.ID})

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"oldWordPl": "pies",
			"oldWordEn": "dog",
			"newWordPl": "kundel",
			"newWordEn": "mongrel",
		},
	}

	result, err := handlers.UpdateTranslation(params)
	assert.NoError(t, err)
	updated, ok := result.(models.Translation)
	assert.True(t, ok)
	assert.Equal(t, newPl.ID, updated.WordIDPl)
	assert.Equal(t, newEn.ID, updated.WordIDEn)
}

func TestDeleteTranslation(t *testing.T) {
	setupTranslationTestDB(t)

	pl := models.Word{Word: "mysz", Language: "pl"}
	en := models.Word{Word: "mouse", Language: "en"}

	utils.DB.Create(&pl)
	utils.DB.Create(&en)
	utils.DB.Create(&models.Translation{WordIDPl: pl.ID, WordIDEn: en.ID})

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"wordPl": "mysz",
			"wordEn": "mouse",
		},
	}

	result, err := handlers.DeleteTranslation(params)
	assert.NoError(t, err)
	deleted, ok := result.(bool)
	assert.True(t, ok)
	assert.True(t, deleted)

	// Check if translation is gone
	var count int64
	utils.DB.Model(&models.Translation{}).
		Where("word_id_pl = ? AND word_id_en = ?", pl.ID, en.ID).
		Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestConcurrentAddTranslation_RaceCondition(t *testing.T) {
	setupTranslationTestDB(t)

	pl := models.Word{Word: "lew", Language: "pl"}
	en := models.Word{Word: "lion", Language: "en"}
	utils.DB.Create(&pl)
	utils.DB.Create(&en)

	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			params := graphql.ResolveParams{
				Args: map[string]interface{}{
					"wordPl": "lew",
					"wordEn": "lion",
				},
			}
			_, _ = handlers.AddTranslation(params)
		}()
	}

	wg.Wait()

	// Powinien istnieć tylko jeden rekord
	var count int64
	err := utils.DB.Model(&models.Translation{}).
		Where("word_id_pl = ? AND word_id_en = ?", pl.ID, en.ID).
		Count(&count).Error

	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Only one translation should exist after concurrent insertions")
}

func TestConcurrentDeleteTranslation_RaceCondition(t *testing.T) {
	setupTranslationTestDB(t)

	pl := models.Word{Word: "wilk", Language: "pl"}
	en := models.Word{Word: "wolf", Language: "en"}

	utils.DB.Create(&pl)
	utils.DB.Create(&en)
	utils.DB.Create(&models.Translation{WordIDPl: pl.ID, WordIDEn: en.ID})

	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			params := graphql.ResolveParams{
				Args: map[string]interface{}{
					"wordPl": "wilk",
					"wordEn": "wolf",
				},
			}
			_, _ = handlers.DeleteTranslation(params)
		}()
	}

	wg.Wait()

	// Check, if translation was deleted correctly
	var count int64
	utils.DB.Model(&models.Translation{}).
		Where("word_id_pl = ? AND word_id_en = ?", pl.ID, en.ID).
		Count(&count)
	assert.Equal(t, int64(0), count, "Translation should be deleted exactly once")
}
