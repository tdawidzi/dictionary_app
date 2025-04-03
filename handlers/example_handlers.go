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

	// Use concurrency to fetch examples at the same time
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

// AddExample adds an example to db associated with given word
func AddExample(p graphql.ResolveParams) (interface{}, error) {
	wordText, _ := p.Args["word"].(string)
	language, _ := p.Args["language"].(string)
	exampleText, _ := p.Args["example"].(string)

	// Search for given word, and retrieve its id from db
	var word models.Word
	if err := utils.DB.Where("word = ? AND language = ?", wordText, language).First(&word).Error; err != nil {
		return nil, fmt.Errorf("word not found: %w", err)
	}

	example := models.Example{
		WordID:  word.ID,
		Example: exampleText,
	}

	// Add example to db
	if err := utils.DB.Create(&example).Error; err != nil {
		return nil, fmt.Errorf("failed to create example: %w", err)
	}

	return example, nil
}

// UpdateExample - modifies example text with given example id
func UpdateExample(p graphql.ResolveParams) (interface{}, error) {
	// Check for errors in id and convert it to integer
	id, ok := p.Args["id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid or missing ID")
	}

	// Check if argument example is given
	newExample, hasExample := p.Args["example"].(string)

	// Find example with given id
	var example models.Example
	if err := utils.DB.First(&example, id).Error; err != nil {
		return nil, fmt.Errorf("example not found: %w", err)
	}

	// Update example if new example is not nil
	if hasExample {
		example.Example = newExample
	}

	// Save changes
	if err := utils.DB.Save(&example).Error; err != nil {
		return nil, fmt.Errorf("failed to update example: %w", err)
	}

	return example, nil
}

// Deleting example with given id from db
func DeleteExample(p graphql.ResolveParams) (interface{}, error) {
	id, ok := p.Args["id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid or missing ID")
	}

	// Find example with given id
	var example models.Example
	if err := utils.DB.First(&example, id).Error; err != nil {
		return false, fmt.Errorf("example not found: %w", err)
	}

	if err := utils.DB.Delete(&example).Error; err != nil {
		return false, fmt.Errorf("failed to delete example: %w", err)
	}

	// Return true if succeeded
	return true, nil
}
