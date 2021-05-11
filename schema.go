package gql

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type SchemaConfig struct {
	Types []reflect.Type
}

func NewSchema(config SchemaConfig) graphql.Schema {
	gTypes := config.Types

	var query *graphql.Object
	var mutation *graphql.Object
	gObjects := []graphql.Type{}

	for _, gType := range gTypes {
		gObject := NewObject(gType)

		switch gObject.Name() {
			case "Query":
				query = gObject
			case "Mutation":
				mutation = gObject
			default:
				gObjects = append(gObjects, gObject)
		}
	}

	schemaConfig := graphql.SchemaConfig{Query: query, Mutation: mutation, Types: gObjects}
	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		panic(err)
	}

	return schema
}