package handlers

import (
	"fmt"
	"sync"

	"github.com/tdawidzi/dictionary_app/models"
	"github.com/tdawidzi/dictionary_app/utils"

	"github.com/graphql-go/graphql"
)

// GetExamplesForWord fetches example sentences for a given word.
func GetExamplesForWord(p graphql.ResolveParams) (interface{}, error) {
	wordText, ok := p.Args["word"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid word input")
	}

	// Retrieve the word from the database
	var word models.Word
	err := utils.DB.Where("word = ?", wordText).First(&word).Error
	if err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	// Fetch all example IDs associated with the word
	var exampleIDs []int
	err = utils.DB.Model(&models.Example{}).Where("word_id = ?", word.ID).Pluck("id", &exampleIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch example IDs: %w", err)
	}

	// Use goroutines to fetch each example concurrently
	var wg sync.WaitGroup
	examplesChan := make(chan models.Example, len(exampleIDs))
	errorsChan := make(chan error, len(exampleIDs))

	for _, exampleID := range exampleIDs {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			var example models.Example
			err := utils.DB.First(&example, id).Error
			if err != nil {
				errorsChan <- fmt.Errorf("failed to fetch example: %w", err)
				return
			}
			examplesChan <- example
		}(exampleID)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(examplesChan)
		close(errorsChan)
	}()

	// Collect results
	var examples []models.Example
	for example := range examplesChan {
		examples = append(examples, example)
	}

	// Check for errors
	if len(errorsChan) > 0 {
		return nil, fmt.Errorf("some examples could not be fetched")
	}

	return examples, nil
}

func AddExample(p graphql.ResolveParams) (interface{}, error) {
	wordText, _ := p.Args["word"].(string)
	language, _ := p.Args["language"].(string)
	exampleText, _ := p.Args["example"].(string)

	var word models.Word
	if err := utils.DB.Where("word = ? AND language = ?", wordText, language).First(&word).Error; err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	example := models.Example{
		WordID:  word.ID,
		Example: exampleText,
	}

	if err := utils.DB.Create(&example).Error; err != nil {
		return nil, fmt.Errorf("failed to create example: %w", err)
	}

	return example, nil
}

func UpdateExample(p graphql.ResolveParams) (interface{}, error) {
	// Sprawdzamy, czy ID istnieje i poprawnie konwertujemy na int
	id, ok := p.Args["id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid or missing ID")
	}

	// Sprawdzamy, czy example został przekazany
	newExample, hasExample := p.Args["example"].(string)

	var example models.Example
	if err := utils.DB.First(&example, id).Error; err != nil {
		return nil, fmt.Errorf("example not found: %w", err)
	}

	if hasExample { // Jeśli podano nowy przykład, aktualizujemy
		example.Example = newExample
	}

	if err := utils.DB.Save(&example).Error; err != nil {
		return nil, fmt.Errorf("failed to update example: %w", err)
	}

	return example, nil
}

func DeleteExample(p graphql.ResolveParams) (interface{}, error) {
	id, ok := p.Args["id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid or missing ID")
	}

	var example models.Example
	if err := utils.DB.First(&example, id).Error; err != nil {
		return false, fmt.Errorf("example not found: %w", err)
	}

	if err := utils.DB.Delete(&example).Error; err != nil {
		return false, fmt.Errorf("failed to delete example: %w", err)
	}

	return true, nil // Zwracamy true, jeśli operacja się powiodła
}
