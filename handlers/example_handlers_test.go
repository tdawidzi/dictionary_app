package handlers_test

import (
	"sync"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/tdawidzi/dictionary_app/handlers"
	"github.com/tdawidzi/dictionary_app/models"
	"github.com/tdawidzi/dictionary_app/testresources"
	"github.com/tdawidzi/dictionary_app/utils"
)

func setupExampleTestDB(t *testing.T) {
	utils.DB = testresources.NewSingleTestConnection(t)
	err := utils.DB.AutoMigrate(&models.Word{}, &models.Example{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}

func TestAddAndGetExample(t *testing.T) {
	setupExampleTestDB(t)

	word := models.Word{Word: "kot", Language: "pl"}
	utils.DB.Create(&word)

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word":     "kot",
			"language": "pl",
			"example":  "Kot siedzi na dachu.",
		},
	}

	result, err := handlers.AddExample(params)
	assert.NoError(t, err)

	example, ok := result.(models.Example)
	assert.True(t, ok)
	assert.Equal(t, "Kot siedzi na dachu.", example.Example)

	// Get examples
	getParams := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word": "kot",
		},
	}
	result, err = handlers.GetExamplesForWord(getParams)
	assert.NoError(t, err)

	examples, ok := result.([]models.Example)
	assert.True(t, ok)
	assert.Len(t, examples, 1)
	assert.Equal(t, "Kot siedzi na dachu.", examples[0].Example)
}

func TestUpdateExample(t *testing.T) {
	setupExampleTestDB(t)

	word := models.Word{Word: "pies", Language: "pl"}
	utils.DB.Create(&word)

	ex := models.Example{WordID: word.ID, Example: "Pies szczeka."}
	utils.DB.Create(&ex)

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id":      int(ex.ID),
			"example": "Pies głośno szczeka.",
		},
	}

	result, err := handlers.UpdateExample(params)
	assert.NoError(t, err)

	updated, ok := result.(models.Example)
	assert.True(t, ok)
	assert.Equal(t, "Pies głośno szczeka.", updated.Example)
}

func TestDeleteExample(t *testing.T) {
	setupExampleTestDB(t)

	word := models.Word{Word: "mysz", Language: "pl"}
	utils.DB.Create(&word)

	ex := models.Example{WordID: word.ID, Example: "Mysz uciekła do nory."}
	utils.DB.Create(&ex)

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id": int(ex.ID),
		},
	}

	result, err := handlers.DeleteExample(params)
	assert.NoError(t, err)

	deleted, ok := result.(bool)
	assert.True(t, ok)
	assert.True(t, deleted)

	// Confirm deletion
	var count int64
	utils.DB.Model(&models.Example{}).Where("id = ?", ex.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestConcurrentAddExample_RaceCondition(t *testing.T) {
	utils.DB = testresources.NewSingleTestConnection(t)
	err := utils.DB.AutoMigrate(&models.Word{}, &models.Example{})
	assert.NoError(t, err)

	// Dodaj słowo, do którego będą dodawane przykłady
	word := models.Word{Word: "testowy", Language: "pl"}
	utils.DB.Create(&word)

	var wg sync.WaitGroup
	concurrency := 10
	exampleText := "To jest przykład wyścigu."

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			params := graphql.ResolveParams{
				Args: map[string]interface{}{
					"word":     word.Word,
					"language": word.Language,
					"example":  exampleText,
				},
			}

			_, _ = handlers.AddExample(params)
		}()
	}

	wg.Wait()

	// Check if only one example exists in db
	var examples []models.Example
	err = utils.DB.Where("example = ? AND word_id = ?", exampleText, word.ID).Find(&examples).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, len(examples), "Should only have one example after concurrent insertions")
}

func TestConcurrentUpdateExample_RaceCondition(t *testing.T) {
	utils.DB = testresources.NewSingleTestConnection(t)
	err := utils.DB.AutoMigrate(&models.Word{}, &models.Example{})
	assert.NoError(t, err)

	// Prepare data
	word := models.Word{Word: "testing", Language: "en"}
	utils.DB.Create(&word)

	example := models.Example{WordID: word.ID, Example: "Original text"}
	utils.DB.Create(&example)

	var wg sync.WaitGroup
	concurrency := 5
	newTexts := []string{
		"Update A", "Update B", "Update C", "Update D", "Update E",
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(text string) {
			defer wg.Done()

			params := graphql.ResolveParams{
				Args: map[string]interface{}{
					"id":      int(example.ID),
					"example": text,
				},
			}

			_, _ = handlers.UpdateExample(params)
		}(newTexts[i])
	}

	wg.Wait()

	// Check final version
	var final models.Example
	err = utils.DB.First(&final, example.ID).Error
	assert.NoError(t, err)

	found := false
	for _, txt := range newTexts {
		if final.Example == txt {
			found = true
			break
		}
	}
	assert.True(t, found, "Final text should be one of these goroutines")
}

func TestDeleteAndUpdateExample_RaceCondition(t *testing.T) {
	utils.DB = testresources.NewSingleTestConnection(t)
	err := utils.DB.AutoMigrate(&models.Word{}, &models.Example{})
	assert.NoError(t, err)

	// Prepare data
	word := models.Word{Word: "testing", Language: "en"}
	utils.DB.Create(&word)

	example := models.Example{WordID: word.ID, Example: "To be deleted"}
	utils.DB.Create(&example)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // delay, to increase collision risk

		params := graphql.ResolveParams{
			Args: map[string]interface{}{
				"id": int(example.ID),
			},
		}
		_, _ = handlers.DeleteExample(params)
	}()

	go func() {
		defer wg.Done()

		params := graphql.ResolveParams{
			Args: map[string]interface{}{
				"id":      int(example.ID),
				"example": "Updated text",
			},
		}
		_, _ = handlers.UpdateExample(params)
	}()

	wg.Wait()

	// Check if record still exists
	var count int64
	utils.DB.Model(&models.Example{}).Where("id = ?", example.ID).Count(&count)

	// Can be 0 or 1, but if 1 exist, should have changed text
	if count == 1 {
		var updated models.Example
		utils.DB.First(&updated, example.ID)
		assert.Equal(t, "Updated text", updated.Example)
	} else {
		assert.Equal(t, int64(0), count)
	}
}
