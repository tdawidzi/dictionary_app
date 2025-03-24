package schema

import (
	"dictionary-app/handlers"

	"github.com/graphql-go/graphql"
)

// Representation of GraphQL schema of app
var Schema *graphql.Schema

// Define types:
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

// var translationType = graphql.NewObject(graphql.ObjectConfig{
// 	Name: "Translation",
// 	Fields: graphql.Fields{
// 		"id": &graphql.Field{
// 			Type: graphql.Int,
// 		},
// 		"id_pl": &graphql.Field{
// 			Type: graphql.Int,
// 		},
// 		"id_en": &graphql.Field{
// 			Type: graphql.Int,
// 		},
// 	},
// })

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
