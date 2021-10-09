package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type Array struct {
	value       GrootType
	parserArray *parser.Array
}

func NewArray(t *parser.Array, builder *SchemaBuilder) (*Array, error) {
	element := t.Element()
	value, err := getOrCreateType(element, builder)
	if err != nil {
		return nil, err
	}

	return &Array{value, t}, nil
}

func (array *Array) GraphQLType() graphql.Type {
	return graphql.NewList(array.value.GraphQLType())
}

func (array *Array) ParserType() parser.Type {
	return array.parserArray
}
