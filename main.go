package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tdawidzi/dictionary_app/config"
	"github.com/tdawidzi/dictionary_app/schema"
	"github.com/tdawidzi/dictionary_app/utils"

	"github.com/graphql-go/graphql"
)

func main() {
	// 1️⃣ Wczytanie konfiguracji
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Błąd ładowania konfiguracji: %v", err)
	}

	// 2️⃣ Połączenie z bazą danych
	err = utils.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Błąd połączenia z bazą danych: %v", err)
	}

	// 3️⃣ Zamknięcie bazy danych po zakończeniu działania aplikacji
	sqlDB, err := utils.DB.DB()
	if err != nil {
		log.Fatalf("Błąd pobrania instancji bazy danych: %v", err)
	}
	defer sqlDB.Close()

	// 4️⃣ Handler GraphQL obsługujący zapytania
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		// Parsowanie zapytania GraphQL z body (dla POST)
		var requestBody struct {
			Query string `json:"query"`
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Błąd odczytu zapytania", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(w, "Błąd parsowania JSON", http.StatusBadRequest)
			return
		}

		// Wykonanie zapytania GraphQL
		params := graphql.Params{
			Schema:        *schema.Schema,
			Context:       context.Background(),
			RequestString: requestBody.Query,
		}
		result := graphql.Do(params)

		// Obsługa błędów GraphQL
		if result.HasErrors() {
			http.Error(w, fmt.Sprintf("Błąd GraphQL: %v", result.Errors), http.StatusInternalServerError)
			return
		}

		// Konwersja wyniku na JSON i zwrócenie odpowiedzi
		response, err := json.Marshal(result)
		if err != nil {
			http.Error(w, "Błąd serializacji wyniku", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})

	// 5️⃣ Uruchomienie serwera
	fmt.Println("Serwer działa na: http://localhost:8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
