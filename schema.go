package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type GrootType interface {
	GraphQLType() graphql.Type
}

type SchemaConfig struct {
	Query      reflect.Type
	Mutation   reflect.Type
	Extensions []graphql.Extension
}

type SchemaBuilder struct {
	types      map[string]graphql.Type
	grootTypes map[reflect.Type]GrootType
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		types:      map[string]graphql.Type{},
		grootTypes: map[reflect.Type]GrootType{},
	}
}

func (builder *SchemaBuilder) parseAndGetRoot(t reflect.Type) *graphql.Object {
	root := NewObject(t, builder)
	return root.GraphQLType().(*graphql.Object)
}

func NewSchema(config SchemaConfig) graphql.Schema {
	builder := NewSchemaBuilder()
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query: builder.parseAndGetRoot(config.Query),
		// Mutation:   builder.parseAndGetRoot(config.Mutation),
		Extensions: config.Extensions,
	})

	return schema
}
