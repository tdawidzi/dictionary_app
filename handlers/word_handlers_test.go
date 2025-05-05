package handlers_test

import (
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
