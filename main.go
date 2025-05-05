package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tdawidzi/dictionary_app/config"
	"github.com/tdawidzi/dictionary_app/schema"
	"github.com/tdawidzi/dictionary_app/utils"

	"github.com/graphql-go/graphql"
)

func main() {
	// Load config
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Error while loading configuration: %v", err)
	}

	// Connect do database
	err = utils.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Error while connecting do database: %v", err)
	}

	// Close DB at the end
	sqlDB, err := utils.DB.DB()
	if err != nil {
		log.Fatalf("Error while loading db instance: %v", err)
	}
	defer sqlDB.Close()

	// GraphQL handler for queries
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		// GraphQL parsing
		var requestBody struct {
			Query string `json:"query"`
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading query", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(w, "Error JSON parsing", http.StatusBadRequest)
			return
		}

		// GraphQL query execution
		params := graphql.Params{
			Schema:        *schema.Schema,
			Context:       context.Background(),
			RequestString: requestBody.Query,
		}
		result := graphql.Do(params)

		// GraphqL Error handling
		if result.HasErrors() {
			http.Error(w, fmt.Sprintf("GraphQL Error: %v", result.Errors), http.StatusInternalServerError)
			return
		}

		// Convert to JSON and return
		response, err := json.Marshal(result)
		if err != nil {
			http.Error(w, "Output serialization Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})

	// Server startup
	fmt.Println("Server listening on: http://localhost:8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
