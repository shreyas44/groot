package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

var (
	_ GrootType = (*Object)(nil)
	_ GrootType = (*InputObject)(nil)
	_ GrootType = (*Interface)(nil)
	_ GrootType = (*Union)(nil)
	_ GrootType = (*Scalar)(nil)
	_ GrootType = (*Enum)(nil)
	_ GrootType = (*Array)(nil)
	_ GrootType = (*NonNull)(nil)
)

type GrootType interface {
	GraphQLType() graphql.Type
	ParserType() parser.Type
}

type SchemaConfig struct {
	Query        *parser.Object
	Mutation     *parser.Object
	Subscription *parser.Object
	Types        []parser.Type
	Extensions   []graphql.Extension
}

type SchemaBuilder struct {
	grootTypes      map[parser.Type]GrootType
	reflectGrootMap map[reflect.Type]GrootType
}

func (builder *SchemaBuilder) addType(t parser.Type, grootType GrootType) {
	builder.grootTypes[t] = grootType
	builder.reflectGrootMap[t.ReflectType()] = grootType
}

func (builder *SchemaBuilder) getType(t parser.Type) (GrootType, bool) {
	grootType, ok := builder.grootTypes[t]
	if ok {
		return grootType, true
	}

	return nil, false
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		grootTypes:      map[parser.Type]GrootType{},
		reflectGrootMap: map[reflect.Type]GrootType{},
	}
}

func NewSchema(config SchemaConfig) (graphql.Schema, error) {
	builder := NewSchemaBuilder()
	schemaConfig := graphql.SchemaConfig{
		Extensions: config.Extensions,
		Types:      []graphql.Type{},
	}

	if config.Query != nil {
		query, err := builder.parseAndGetRoot(config.Query)
		if err != nil {
			return graphql.Schema{}, err
		}

		schemaConfig.Query = query
	}

	if config.Mutation != nil {
		mutation, err := builder.parseAndGetRoot(config.Mutation)
		if err != nil {
			return graphql.Schema{}, err
		}

		schemaConfig.Mutation = mutation
	}

	if config.Subscription != nil {
		subscription, err := builder.parseAndGetRoot(config.Subscription)
		if err != nil {
			return graphql.Schema{}, err
		}

		schemaConfig.Subscription = subscription
	}

	for _, t := range config.Types {
		grootType, err := getOrCreateType(t, builder)
		if err != nil {
			return graphql.Schema{}, err
		}

		schemaConfig.Types = append(schemaConfig.Types, grootType.GraphQLType())
	}

	return graphql.NewSchema(schemaConfig)
}

func (builder *SchemaBuilder) parseAndGetRoot(t *parser.Object) (*graphql.Object, error) {
	root, err := NewObject(t, builder)
	if err != nil {
		return nil, err
	}

	return root.GraphQLType().(*graphql.Object), nil
}
