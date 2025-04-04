package schema

import (
	//"sync"
	"testing"

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

// 🟢 Test GetWords - Pobieranie wszystkich słów
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

// 🟢 Test AddWord - Dodawanie nowego słowa
func TestAddWord(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordText := "sun"
	language := "en"

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "words" \("word","language"\) VALUES \(\$1,\$2\) RETURNING "id"`).
		WithArgs(wordText, language).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // Zwracamy ID nowo dodanego słowa
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

// 🟢 Test UpdateWord - Modyfikacja istniejącego słowa
func TestUpdateWord(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	oldWord := "sun"
	newWord := "house"
	language := "en"

	// Mockowanie wyszukiwania słowa
	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = \$2 ORDER BY "words"."id" LIMIT \$3`).
		WithArgs(oldWord, language, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, oldWord, language))

	// Mockowanie aktualizacji słowa
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "words" SET "word"=\$1,"language"=\$2 WHERE "id" = \$3`).
		WithArgs(newWord, language, 1).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Poprawka
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

// 🟢 Test DeleteWord - Usuwanie słowa
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

// Test dla GetExamplesForWord
func TestGetExamplesForWord(t *testing.T) {
	// Tworzymy mock bazy danych SQL
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer sqlDB.Close()

	// Tworzymy instancję GORM z mockiem SQL
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	assert.NoError(t, err)

	// Podmieniamy globalne utils.DB na mockowaną bazę danych
	utils.DB = gormDB

	// Dane testowe
	wordText := "test"
	word := models.Word{ID: 1, Word: wordText}
	exampleIDs := []int{1, 2}
	example1 := models.Example{ID: 1, WordID: 1, Example: "Example 1"}
	example2 := models.Example{ID: 2, WordID: 1, Example: "Example 2"}

	// Oczekiwane zapytania do bazy danych

	// Pobranie słowa z bazy danych
	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordText, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word"}).AddRow(1, "test"))

	// Pobranie listy przykładów
	mock.ExpectQuery(`SELECT "id" FROM "examples" WHERE word_id = \$1`).
		WithArgs(word.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(exampleIDs[0]).AddRow(exampleIDs[1]))

	// Pobranie treści przykładu o ID 1
	mock.ExpectQuery(`SELECT \* FROM "examples" WHERE "examples"."id" = \$1 ORDER BY "examples"."id" LIMIT \$2`).
		WithArgs(example1.ID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).AddRow(example1.ID, example1.WordID, example1.Example))

	// Pobranie treści przykładu o ID 2
	mock.ExpectQuery(`SELECT \* FROM "examples" WHERE "examples"."id" = \$1 ORDER BY "examples"."id" LIMIT \$2`).
		WithArgs(example2.ID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).AddRow(example2.ID, example2.WordID, example2.Example))

	// Parametry GraphQL
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"word": wordText,
		},
	}

	// Wywołanie funkcji
	result, err := handlers.GetExamplesForWord(params)
	// Obsługa błędu w wyniku result == nil
	if err != nil {
		t.Fatalf("Unexpected error: %v", err) // Teraz test od razu przerywa działanie
	}
	assert.Len(t, result, 2)

	// Sprawdzenie wyników
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Example 1", result.([]models.Example)[0].Example)
	assert.Equal(t, "Example 2", result.([]models.Example)[1].Example)

	// Sprawdzenie, czy wszystkie zapytania zostały wykonane
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

func TestAddExample(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordText := "test"
	language := "en"
	exampleText := "This is an example"

	// Mockowanie zapytania do znalezienia słowa
	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = \$2 ORDER BY "words"."id"`).
		WithArgs(wordText, language, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, wordText, language))

	// Mockowanie wstawienia przykładu do bazy
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

func TestUpdateExample(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	exampleID := 1
	newExampleText := "Updated example"

	// Mockowanie pobrania istniejącego przykładu
	mock.ExpectQuery(`SELECT \* FROM "examples" WHERE "examples"."id" = \$1`).
		WithArgs(exampleID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).AddRow(exampleID, 1, "Old example"))

	// Mockowanie aktualizacji rekordu
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

func TestDeleteExample(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	exampleID := 1

	// Mockowanie pobrania przykładu przed usunięciem
	mock.ExpectQuery(`SELECT \* FROM "examples" WHERE "examples"."id" = \$1`).
		WithArgs(exampleID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "example"}).AddRow(exampleID, 1, "Some example"))

	// Mockowanie usunięcia przykładu
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

func TestAddTranslation(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordPl := "dom"
	wordEn := "house"

	// Mockowanie zapytań o istnienie słów
	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'pl' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordPl, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, wordPl, "pl"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'en' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordEn, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(2, wordEn, "en"))

	// Mockowanie dodania tłumaczenia
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

	assert.NoError(t, mock.ExpectationsWereMet()) // Sprawdzenie wywołań
}

func TestUpdateTranslation(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	oldWordPl, newWordPl := "dom", "budynek"
	oldWordEn, newWordEn := "house", "building"

	// Mockowanie zapytań dla starych słów
	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'pl' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(oldWordPl, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, oldWordPl, "pl"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'en' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(oldWordEn, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(2, oldWordEn, "en"))

	// Mockowanie zapytań dla nowych słów
	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'pl' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(newWordPl, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(3, newWordPl, "pl"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'en' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(newWordEn, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(4, newWordEn, "en"))

	// Mockowanie pobrania istniejącego tłumaczenia
	mock.ExpectQuery(`SELECT \* FROM "translations" WHERE word_id_pl = \$1 AND word_id_en = \$2 ORDER BY "translations"."id" LIMIT \$3`).
		WithArgs(1, 2, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id_pl", "word_id_en"}).AddRow(1, 1, 2))

	// Mockowanie aktualizacji tłumaczenia
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

	assert.NoError(t, mock.ExpectationsWereMet()) // Sprawdzenie wywołań
}

func TestDeleteTranslation(t *testing.T) {
	mock, teardown := setupTestDB(t)
	defer teardown()

	wordPl := "dom"
	wordEn := "house"

	// Mockowanie zapytań o istnienie słów
	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'pl' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordPl, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(1, wordPl, "pl"))

	mock.ExpectQuery(`SELECT \* FROM "words" WHERE word = \$1 AND language = 'en' ORDER BY "words"."id" LIMIT \$2`).
		WithArgs(wordEn, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "language"}).AddRow(2, wordEn, "en"))

	// Mockowanie usuwania tłumaczenia
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

	assert.NoError(t, mock.ExpectationsWereMet()) // Sprawdzenie wywołań
}
