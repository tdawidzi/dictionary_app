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

func setupTestDB(t *testing.T) {
	utils.DB = testresources.NewSingleTestConnection(t)
	err := utils.DB.AutoMigrate(&models.Word{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}

func TestAddAndGetWord(t *testing.T) {
	setupTestDB(t)

	// Add word
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word":     "kot",
			"language": "pl",
		},
	}

	result, err := handlers.AddWord(params)
	assert.NoError(t, err)

	word, ok := result.(models.Word)
	assert.True(t, ok)
	assert.Equal(t, "kot", word.Word)
	assert.Equal(t, "pl", word.Language)

	// Get word
	getParams := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word": "kot",
		},
	}

	result, err = handlers.GetWordByText(getParams)
	assert.NoError(t, err)

	word, ok = result.(models.Word)
	assert.True(t, ok)
	assert.Equal(t, "kot", word.Word)
}

func TestGetWords(t *testing.T) {
	setupTestDB(t)

	// Insert test words
	utils.DB.Create(&models.Word{Word: "pies", Language: "pl"})
	utils.DB.Create(&models.Word{Word: "dog", Language: "en"})

	params := graphql.ResolveParams{}
	result, err := handlers.GetWords(params)
	assert.NoError(t, err)

	words, ok := result.([]models.Word)
	assert.True(t, ok)
	assert.Len(t, words, 2)
}

func TestUpdateWord(t *testing.T) {
	setupTestDB(t)

	// Add word to update
	utils.DB.Create(&models.Word{Word: "stary", Language: "pl"})

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"oldWord":  "stary",
			"language": "pl",
			"newWord":  "nowy",
		},
	}

	result, err := handlers.UpdateWord(params)
	assert.NoError(t, err)

	updatedWord, ok := result.(models.Word)
	assert.True(t, ok)
	assert.Equal(t, "nowy", updatedWord.Word)
}

func TestDeleteWord(t *testing.T) {
	setupTestDB(t)

	// Add word to delete
	utils.DB.Create(&models.Word{Word: "usun", Language: "pl"})

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word":     "usun",
			"language": "pl",
		},
	}

	result, err := handlers.DeleteWord(params)
	assert.NoError(t, err)

	deleted, ok := result.(bool)
	assert.True(t, ok)
	assert.True(t, deleted)

	// Confirm deletion
	var count int64
	utils.DB.Model(&models.Word{}).Where("word = ?", "usun").Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestAddWordRaceCondition(t *testing.T) {
	setupTestDB(t)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			params := graphql.ResolveParams{
				Args: map[string]interface{}{
					"word":     "kot",
					"language": "pl",
				},
			}
			_, _ = handlers.AddWord(params)
		}()
	}
	wg.Wait()

	// Expect only one entry
	var count int64
	utils.DB.Model(&models.Word{}).Where("word = ? AND language = ?", "kot", "pl").Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestUpdateWordRaceCondition(t *testing.T) {
	setupTestDB(t)

	// Start with a base word
	utils.DB.Create(&models.Word{Word: "dom", Language: "pl"})

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	newWords := []string{"dom1", "dom2", "dom3", "dom4", "dom5", "dom6", "dom7", "dom8", "dom9", "dom10"}

	for i := 0; i < goroutines; i++ {
		go func(newWord string) {
			defer wg.Done()
			params := graphql.ResolveParams{
				Args: map[string]interface{}{
					"oldWord":  "dom",
					"language": "pl",
					"newWord":  newWord,
				},
			}
			_, _ = handlers.UpdateWord(params)
		}(newWords[i])
	}
	wg.Wait()

	// Should be only one record, but its value is last committed one (non-deterministic)
	var words []models.Word
	utils.DB.Find(&words)
	assert.Len(t, words, 1)
	assert.Contains(t, newWords, words[0].Word)
}

func TestDeleteWordRaceCondition(t *testing.T) {
	setupTestDB(t)

	word := models.Word{Word: "usun", Language: "pl"}
	utils.DB.Create(&word)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			params := graphql.ResolveParams{
				Args: map[string]interface{}{
					"word":     "usun",
					"language": "pl",
				},
			}
			_, _ = handlers.DeleteWord(params)
		}()
	}
	wg.Wait()

	var count int64
	utils.DB.Model(&models.Word{}).Where("word = ?", "usun").Count(&count)
	assert.Equal(t, int64(0), count)
}
