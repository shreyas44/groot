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
	Types      []reflect.Type
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

func NewSchema(config SchemaConfig) (graphql.Schema, error) {
	builder := NewSchemaBuilder()
	schemaConfig := graphql.SchemaConfig{
		Extensions: config.Extensions,
		Types:      []graphql.Type{},
	}

	if config.Query != nil {
		schemaConfig.Query = builder.parseAndGetRoot(config.Query)
	}

	if config.Mutation != nil {
		schemaConfig.Mutation = builder.parseAndGetRoot(config.Mutation)
	}

	for _, t := range config.Types {
		schemaConfig.Types = append(schemaConfig.Types, NewObject(t, builder).GraphQLType())
	}

	return graphql.NewSchema(schemaConfig)
}
