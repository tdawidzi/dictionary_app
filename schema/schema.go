package schema

import (
	"github.com/tdawidzi/dictionary_app/handlers"

	"github.com/graphql-go/graphql"
)

// Representation of GraphQL schema of app
var Schema *graphql.Schema

// Define types (from models.go) as graphql objects
var wordType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Word",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"word": &graphql.Field{
			Type: graphql.String,
		},
		"language": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var translationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Translation",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"id_pl": &graphql.Field{
			Type: graphql.Int,
		},
		"id_en": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var exampleType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Example",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"example": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Define queries
var RootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		// Fetch all words
		"words": &graphql.Field{
			Type:    graphql.NewList(wordType),
			Resolve: handlers.GetWords,
		},

		// Fetch translations for a given word
		"translationsForWord": &graphql.Field{
			Type: graphql.NewList(wordType), // Returns a list of translated words
			Args: graphql.FieldConfigArgument{
				"word": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: handlers.GetTranslationsForWord,
		},

		// Fetch examples for a given word
		"examplesForWord": &graphql.Field{
			Type: graphql.NewList(exampleType), // Returns a list of example sentences
			Args: graphql.FieldConfigArgument{
				"word": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: handlers.GetExamplesForWord,
		},
	},
})

// Define Mutations
var RootMutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		// Add a new word
		"addWord": &graphql.Field{
			Type: wordType,
			Args: graphql.FieldConfigArgument{
				"word": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"language": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
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

// Initialization
func init() {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    RootQuery,
		Mutation: RootMutation,
	})
	if err != nil {
		panic(err)
	}
	Schema = &schema
}
