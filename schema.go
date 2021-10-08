package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type GrootType interface {
	GraphQLType() graphql.Type
	ReflectType() reflect.Type
}

type SchemaConfig struct {
	Query        *parser.Type
	Mutation     *parser.Type
	Subscription *parser.Type
	Types        []*parser.Type
	Extensions   []graphql.Extension
}

type SchemaBuilder struct {
	grootTypes      map[*parser.Type]GrootType
	reflectGrootMap map[reflect.Type]GrootType
}

func (builder *SchemaBuilder) addType(t *parser.Type, grootType GrootType) {
	builder.grootTypes[t] = grootType

	if t.Kind() == parser.Interface {
		builder.grootTypes[t.Definition()] = grootType
		builder.reflectGrootMap[t.Definition().Type] = grootType
	}

	builder.reflectGrootMap[t.Type] = grootType
}

func (builder *SchemaBuilder) getType(t *parser.Type) (GrootType, bool) {
	grootType, ok := builder.grootTypes[t]
	if ok {
		return grootType, true
	}

	if !ok && t.Kind() == parser.Interface {
		return builder.getType(t.Definition())
	}

	return nil, false
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		grootTypes:      map[*parser.Type]GrootType{},
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

func (builder *SchemaBuilder) parseAndGetRoot(t *parser.Type) (*graphql.Object, error) {
	root, err := NewObject(t, builder)
	if err != nil {
		return nil, err
	}

	return root.GraphQLType().(*graphql.Object), nil
}
