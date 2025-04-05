package schema

import (
	//"sync"

	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/tdawidzi/dictionary_app/handlers"
	"github.com/tdawidzi/dictionary_app/models"
	"github.com/tdawidzi/dictionary_app/utils"
)

// Test GetWords - Fetchung all words from database
func TestGetWords(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	mock.ExpectQuery(`SELECT \* FROM "words"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
			AddRow(1, "house", "en").
			AddRow(2, "dom", "pl"))

	params := graphql.ResolveParams{}

	result, err := handlers.GetWords(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	words, ok := result.([]models.Word)
	assert.True(t, ok)
	assert.Len(t, words, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test AddWord - Adding new word to database
func TestAddWord(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordText := "sun"
	language := "en"

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "words" \("word","language"\) VALUES \(\$1,\$2\) RETURNING "id"`).
		WithArgs(wordText, language).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // id of newly added word
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word":     wordText,
			"language": language,
		},
	}

	result, err := handlers.AddWord(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	word, ok := result.(models.Word)
	assert.True(t, ok)
	assert.Equal(t, wordText, word.Word)
	assert.Equal(t, language, word.Language)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test UpdateWord - Modification of existing word
func TestUpdateWord(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	oldWord := "sun"
	newWord := "house"
	language := "en"

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = \$2 ORDER BY "words"."id" LIMIT \$3`).
		WithArgs(oldWord, language, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, oldWord, language))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "words" SET "word"=\$1,"language"=\$2 WHERE "id" = \$3`).
		WithArgs(newWord, language, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"oldWord":  oldWord,
			"newWord":  newWord,
			"language": language,
		},
	}

	result, err := handlers.UpdateWord(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	word, ok := result.(models.Word)
	assert.True(t, ok)
	assert.Equal(t, newWord, word.Word)
	assert.Equal(t, language, word.Language)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test DeleteWord - Deletion of word existing in database
func TestDeleteWord(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordText := "house"
	language := "en"

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = \$2`).
		WithArgs(wordText, language, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, wordText, language))

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "words" WHERE "words"."id" = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word":     wordText,
			"language": language,
		},
	}

	result, err := handlers.DeleteWord(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.(bool))

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetExamplesForWord - fetching all examples for given word
// Consists of different test, simulating different possible concurrency vulnerabilities
func TestGetExamplesForWord(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "house"
		word := models.Word{ID: 1, Word: wordText, Language: "en"}

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		mock.ExpectQuery(`SELECT "id" FROM "examples"`).
			WithArgs(word.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10).AddRow(20))

		// Allow any order for concurrent queries
		mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery(`SELECT \* FROM "examples"`).
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).
				AddRow(10, word.ID, "Example 1"))
		mock.ExpectQuery(`SELECT \* FROM "examples"`).
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).
				AddRow(20, word.ID, "Example 2"))

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		result, err := handlers.GetExamplesForWord(params)

		assert.NoError(t, err)
		assert.Len(t, result.([]models.Example), 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no examples found", func(t *testing.T) {
		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "house"
		word := models.Word{ID: 1, Word: wordText, Language: "en"}

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		mock.ExpectQuery(`SELECT "id" FROM "examples"`).
			WithArgs(word.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		result, err := handlers.GetExamplesForWord(params)

		assert.NoError(t, err)
		assert.Len(t, result.([]models.Example), 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("partial failure", func(t *testing.T) {
		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "house"
		word := models.Word{ID: 1, Word: wordText, Language: "en"}

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		mock.ExpectQuery(`SELECT "id" FROM "examples"`).
			WithArgs(word.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10).AddRow(20))

		mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery(`SELECT \* FROM "examples"`).
			WithArgs(10, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).
				AddRow(10, word.ID, "Example 1"))
		mock.ExpectQuery(`SELECT \* FROM "examples"`).
			WithArgs(20, 1).
			WillReturnError(fmt.Errorf("not found"))

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		_, err := handlers.GetExamplesForWord(params)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some examples could not be fetched")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("goroutine leak check", func(t *testing.T) {
		// Store initial goroutine count
		initial := runtime.NumGoroutine()

		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "house"
		word := models.Word{ID: 1, Word: wordText, Language: "en"}

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		mock.ExpectQuery(`SELECT "id" FROM "examples"`).
			WithArgs(word.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

		mock.ExpectQuery(`SELECT \* FROM "examples"`).
			WithArgs(10, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).
				AddRow(10, word.ID, "Example 1"))

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		handlers.GetExamplesForWord(params)

		// Allow some time for goroutines to finish
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, initial, runtime.NumGoroutine(), "goroutine leak detected")
	})
}

// Test AddExample - adding new example to database
func TestAddExample(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordText := "test"
	language := "en"
	exampleText := "This is an example"

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = \$2 ORDER BY "words"."id"`).
		WithArgs(wordText, language, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, wordText, language))

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "examples" \("word_id","example"\) VALUES \(\$1,\$2\) RETURNING "id"`).
		WithArgs(1, exampleText).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word":     wordText,
			"language": language,
			"example":  exampleText,
		},
	}

	result, err := handlers.AddExample(params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, exampleText, result.(models.Example).Example)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test Update example - modify example already existing in database
func TestUpdateExample(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	exampleID := 1
	newExampleText := "Updated example"

	mock.ExpectQuery(`SELECT \* FROM "examples" WHERE "examples"."id" = \$1`).
		WithArgs(exampleID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).AddRow(exampleID, 1, "Old example"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "examples" SET "word_id"=\$1,"example"=\$2 WHERE "id" = \$3`).
		WithArgs(1, "Updated example", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id":      exampleID,
			"example": newExampleText,
		},
	}

	result, err := handlers.UpdateExample(params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newExampleText, result.(models.Example).Example)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test Delete example - deleting example already existing in database
func TestDeleteExample(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	exampleID := 1

	mock.ExpectQuery(`SELECT \* FROM "examples" WHERE "examples"."id" = \$1`).
		WithArgs(exampleID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).AddRow(exampleID, 1, "Some example"))

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "examples" WHERE "examples"."id" = \$1`).
		WithArgs(exampleID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id": exampleID,
		},
	}

	result, err := handlers.DeleteExample(params)
	assert.NoError(t, err)
	assert.True(t, result.(bool))

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetTranslationForWord - Fetching all translations of given word from database
// Consists of different test, simulating different possible concurrency vulnerabilities
func TestGetTranslationsForWord(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "dom"
		word := models.Word{ID: 1, Word: wordText, Language: "pl"}

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		mock.ExpectQuery(`SELECT \* FROM "translations"`).
			WithArgs(word.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_id_pl", "word_id_en"}).
				AddRow(1, 1, 2).AddRow(2, 1, 3))

		mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(2, "house", "en"))
		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(3, "home", "en"))

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		result, err := handlers.GetTranslationsForWord(params)

		assert.NoError(t, err)
		assert.Len(t, result.([]models.Word), 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no translations found", func(t *testing.T) {
		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "dom"
		word := models.Word{ID: 1, Word: wordText, Language: "pl"}

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		mock.ExpectQuery(`SELECT \* FROM "translations"`).
			WithArgs(word.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_id_pl", "word_id_en"}))

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		result, err := handlers.GetTranslationsForWord(params)

		assert.NoError(t, err)
		assert.Len(t, result.([]models.Word), 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("partial translation failure", func(t *testing.T) {
		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "dom"
		word := models.Word{ID: 1, Word: wordText, Language: "pl"}

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		mock.ExpectQuery(`SELECT \* FROM "translations"`).
			WithArgs(word.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_id_pl", "word_id_en"}).
				AddRow(1, 1, 2).AddRow(2, 1, 3))

		mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(2, "house", "en"))
		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(3, 1).
			WillReturnError(fmt.Errorf("not found"))

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		_, err := handlers.GetTranslationsForWord(params)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some translations could not be fetched")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("goroutine leak check", func(t *testing.T) {
		// Get baseline count before any test setup
		initial := runtime.NumGoroutine()

		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "house"
		word := models.Word{ID: 1, Word: wordText, Language: "en"}

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		mock.ExpectQuery(`SELECT "id" FROM "examples"`).
			WithArgs(word.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

		mock.ExpectQuery(`SELECT \* FROM "examples"`).
			WithArgs(10, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).
				AddRow(10, word.ID, "Example 1"))

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		result, err := handlers.GetExamplesForWord(params)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Allow more time for cleanup
		time.Sleep(200 * time.Millisecond)

		// Allow 2 goroutines variance (test framework might create some)
		current := runtime.NumGoroutine()
		if current > initial+2 {
			t.Errorf("goroutine leak detected: expected <= %d, got %d", initial+2, current)
		}
	})

	t.Run("unsupported language", func(t *testing.T) {
		mock, teardown := setupTestDB(t)
		defer teardown()

		wordText := "dom"
		word := models.Word{ID: 1, Word: wordText, Language: "de"} // Unsupported

		mock.ExpectQuery(`SELECT \* FROM "words"`).
			WithArgs(wordText, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).
				AddRow(word.ID, word.Word, word.Language))

		// No expectations for further queries since we should return early

		params := graphql.ResolveParams{Args: map[string]interface{}{"word": wordText}}
		_, err := handlers.GetTranslationsForWord(params)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported language")
		assert.NoError(t, mock.ExpectationsWereMet()) // Verify no unexpected queries were made
	})
}

// Test AddTranslation - adding new translation to database
func TestAddTranslation(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordPl := "dom"
	wordEn := "house"

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'pl' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordPl, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, wordPl, "pl"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'en' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordEn, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(2, wordEn, "en"))

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "translations" \("word_id_pl","word_id_en"\) VALUES \(\$1,\$2\) RETURNING "id"`).
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"wordPl": wordPl,
			"wordEn": wordEn,
		},
	}

	result, err := handlers.AddTranslation(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	translation, ok := result.(models.Translation)
	assert.True(t, ok)
	assert.Equal(t, uint(1), translation.WordIDPl)
	assert.Equal(t, uint(2), translation.WordIDEn)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test Update Translation - Modify translation that already exists in dtatabase
func TestUpdateTranslation(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	oldWordPl, newWordPl := "dom", "budynek"
	oldWordEn, newWordEn := "house", "building"

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'pl' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(oldWordPl, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, oldWordPl, "pl"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'en' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(oldWordEn, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(2, oldWordEn, "en"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'pl' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(newWordPl, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(3, newWordPl, "pl"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'en' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(newWordEn, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(4, newWordEn, "en"))

	mock.ExpectQuery(`SELECT \* FROM "translations" WHERE word_id_pl = \$1 AND word_id_en = \$2 ORDER BY "translations"."id" LIMIT \$3`).
		WithArgs(1, 2, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id_pl", "word_id_en"}).AddRow(1, 1, 2))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "translations"`).
		WithArgs(3, 4, 1). // WordIDPl, WordIDEn, ID
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"oldWordPl": oldWordPl,
			"oldWordEn": oldWordEn,
			"newWordPl": newWordPl,
			"newWordEn": newWordEn,
		},
	}

	result, err := handlers.UpdateTranslation(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	translation, ok := result.(models.Translation)
	assert.True(t, ok)
	assert.Equal(t, 3, int(translation.WordIDPl))
	assert.Equal(t, 4, int(translation.WordIDEn))

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test DeleteTranslation - delete translation already existing in database
func TestDeleteTranslation(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordPl := "dom"
	wordEn := "house"

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'pl' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordPl, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, wordPl, "pl"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'en' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordEn, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(2, wordEn, "en"))

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "translations" WHERE word_id_pl = \$1 AND word_id_en = \$2`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"wordPl": wordPl,
			"wordEn": wordEn,
		},
	}

	result, err := handlers.DeleteTranslation(params)

	assert.NoError(t, err)
	assert.True(t, result.(bool))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func setupTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	assert.NoError(t, err)

	utils.DB = gormDB

	return mock, func() {
		sqlDB.Close()
	}
}
