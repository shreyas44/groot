package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type GrootType interface {
	GraphQLType() graphql.Type
	ReflectType() reflect.Type
}

type SchemaConfig struct {
	Query        reflect.Type
	Mutation     reflect.Type
	Subscription reflect.Type
	Types        []reflect.Type
	Extensions   []graphql.Extension
}

type SchemaBuilder struct {
	grootTypes map[reflect.Type]GrootType
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		grootTypes: map[reflect.Type]GrootType{},
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
		if _, ok := builder.grootTypes[t]; !ok {
			object, err := NewObject(t, builder)
			if err != nil {
				return graphql.Schema{}, err
			}

			schemaConfig.Types = append(schemaConfig.Types, object.GraphQLType())
		}
	}

	return graphql.NewSchema(schemaConfig)
}

func (builder *SchemaBuilder) getType(t reflect.Type) GrootType {
	parserType, err := getParserType(t)
	if err != nil {
		return nil
	}

	switch parserType {
	case ParserCustomScalar:
		t = reflect.PtrTo(t)
	case ParserInterface:
		t = t.Method(0).Type.Out(0)
	}

	if grootType, ok := builder.grootTypes[t]; ok {
		return grootType
	}

	return nil
}

func (builder *SchemaBuilder) parseAndGetRoot(t reflect.Type) (*graphql.Object, error) {
	root, err := NewObject(t, builder)
	if err != nil {
		return nil, err
	}

	return root.GraphQLType().(*graphql.Object), nil
}
