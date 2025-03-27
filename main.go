package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tdawidzi/dictionary_app/config"
	"github.com/tdawidzi/dictionary_app/schema"
	"github.com/tdawidzi/dictionary_app/utils"

	"github.com/graphql-go/graphql"
)

func main() {
	config, err := config.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = utils.ConnectDB(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer utils.DB.Close()

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		params := graphql.Params{
			Schema: *schema.Schema,
			// Request:       r,
			Context:       context.Background(),
			OperationName: r.URL.Query().Get("query"),
		}
		result := graphql.Do(params)
		if result.HasErrors() {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %s", result.Errors)))
			return
		}
		w.WriteHeader(http.StatusOK)

		data, err := json.Marshal(result.Data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error marshalling result data: %s", err)))
			return
		}

		w.Write(data)
	})
	fmt.Println("Server listening on http://localhost:8080/graphql")
	http.ListenAndServe(":8080", nil)
}
