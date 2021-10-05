package groot

import "github.com/graphql-go/graphql"

type Array struct {
	value GrootType
}

func (list *Array) GraphQLType() graphql.Type {
	return graphql.NewList(list.value.GraphQLType())
}

func NewArray(value GrootType) *Array {
	return &Array{value: value}
}
