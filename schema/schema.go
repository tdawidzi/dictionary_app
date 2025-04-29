package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/tdawidzi/dictionary_app/handlers"
)

var Schema *graphql.Schema

// Zadeklaruj zmienne typów
var wordType *graphql.Object
var translationType *graphql.Object
var exampleType *graphql.Object

func init() {
	initTypes()

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    buildRootQuery(),
		Mutation: buildRootMutation(),
	})
	if err != nil {
		panic(err)
	}
	Schema = &schema
}

func initTypes() {
	wordType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Word",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.Int,
				},
				"word": &graphql.Field{
					Type: graphql.String,
				},
				"language": &graphql.Field{
					Type: graphql.String,
				},
				"translations": &graphql.Field{
					Type:    graphql.NewList(wordType),
					Resolve: handlers.GetTranslationsForWord,
				},
			}
		}),
	})

	translationType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Translation",
		Fields: graphql.Fields{
			"id":    &graphql.Field{Type: graphql.Int},
			"id_pl": &graphql.Field{Type: graphql.Int},
			"id_en": &graphql.Field{Type: graphql.Int},
		},
	})

	exampleType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Example",
		Fields: graphql.Fields{
			"id":      &graphql.Field{Type: graphql.Int},
			"example": &graphql.Field{Type: graphql.String},
		},
	})
}

func buildRootQuery() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"words": &graphql.Field{
				Type:    graphql.NewList(wordType),
				Resolve: handlers.GetWords,
			},
			"examplesForWord": &graphql.Field{
				Type: graphql.NewList(exampleType),
				Args: graphql.FieldConfigArgument{
					"word": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: handlers.GetExamplesForWord,
			},
			"word": &graphql.Field{
				Type: wordType,
				Args: graphql.FieldConfigArgument{
					"word": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: handlers.GetWordByText,
			},
		},
	})
}

func buildRootMutation() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			// Add a new word
			"addWord": &graphql.Field{
				Type: wordType,
				Args: graphql.FieldConfigArgument{
					"word":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"language": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: handlers.AddWord,
			},
			// Update an existing word
			"updateWord": &graphql.Field{
				Type: wordType,
				Args: graphql.FieldConfigArgument{
					"oldWord": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"newWord": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"language": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: handlers.UpdateWord,
			},

			// Delete a word
			"deleteWord": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"word": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"language": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: handlers.DeleteWord,
			},

			// Add a new translation
			"addTranslation": &graphql.Field{
				Type: translationType,
				Args: graphql.FieldConfigArgument{
					"wordPl": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"wordEn": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: handlers.AddTranslation,
			},

			// Update an existing translation
			"updateTranslation": &graphql.Field{
				Type: translationType,
				Args: graphql.FieldConfigArgument{
					"oldWordPl": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"oldWordEn": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"newWordPl": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"newWordEn": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: handlers.UpdateTranslation,
			},

			// Delete a translation
			"deleteTranslation": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"wordPl": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"wordEn": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: handlers.DeleteTranslation,
			},

			// Add a new example
			"addExample": &graphql.Field{
				Type: exampleType,
				Args: graphql.FieldConfigArgument{
					"word": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"language": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"example": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: handlers.AddExample,
			},

			// Update an existing example
			"updateExample": &graphql.Field{
				Type: exampleType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"example": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: handlers.UpdateExample,
			},

			// Delete an example
			"deleteExample": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: handlers.DeleteExample,
			},
		},
	})
}
