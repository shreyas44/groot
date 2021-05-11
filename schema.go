package gql

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type SchemaConfig struct {
	Types []reflect.Type
	CamelCase *bool
}

func NewSchema(config SchemaConfig) graphql.Schema {
	types := config.Types

	var query *graphql.Object
	var mutation *graphql.Object
	gObjects := []graphql.Type{}
	relationQueue := RelationQueue{items: []queueItem{}}

	for _, gType := range types {
		if config.CamelCase == nil {
			value := true
			config.CamelCase = &value
		}

		objectConfig := ObjectConfig{CamelCase: *config.CamelCase}
		gObject := NewObject(gType, &relationQueue, objectConfig)

		switch gObject.Name() {
			case "Query":
				query = gObject
			case "Mutation":
				mutation = gObject
			default:
				gObjects = append(gObjects, gObject)
		}
	}

	relationQueue.dispatch()

	schemaConfig := graphql.SchemaConfig{Query: query, Mutation: mutation, Types: gObjects}
	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		panic(err)
	}

	return schema
}