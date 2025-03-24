package handlers

import (
	"dictionary-app/models"
	"dictionary-app/utils"
	"fmt"
	"sync"

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
