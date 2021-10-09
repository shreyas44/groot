package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type SchemaConfig struct {
	Query        *parser.Object
	Mutation     *parser.Object
	Subscription *parser.Object
	Types        []parser.Type
	Extensions   []graphql.Extension
}

type SchemaBuilder struct {
	graphqlTypes    map[parser.Type]graphql.Type
	reflectGrootMap map[reflect.Type]graphql.Type
}

func (builder *SchemaBuilder) addType(t parser.Type, graphqlType graphql.Type) {
	builder.graphqlTypes[t] = graphqlType
	builder.reflectGrootMap[t.ReflectType()] = graphqlType
}

func (builder *SchemaBuilder) getType(t parser.Type) (graphql.Type, bool) {
	graphqlType, ok := builder.graphqlTypes[t]
	if ok {
		return graphqlType, true
	}

	return nil, false
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		graphqlTypes:    map[parser.Type]graphql.Type{},
		reflectGrootMap: map[reflect.Type]graphql.Type{},
	}
}

func NewSchema(config SchemaConfig) (graphql.Schema, error) {
	builder := NewSchemaBuilder()
	schemaConfig := graphql.SchemaConfig{
		Extensions: config.Extensions,
		Types:      []graphql.Type{},
	}

	if config.Query != nil {
		schemaConfig.Query = NewObject(config.Query, builder)
	}

	if config.Mutation != nil {
		schemaConfig.Mutation = NewObject(config.Mutation, builder)
	}

	if config.Subscription != nil {
		schemaConfig.Subscription = NewObject(config.Subscription, builder)
	}

	for _, t := range config.Types {
		schemaConfig.Types = append(schemaConfig.Types, getOrCreateType(t, builder))
	}

	return graphql.NewSchema(schemaConfig)
}
