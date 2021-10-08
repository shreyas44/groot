package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type Array struct {
	value GrootType
}

func NewArray(value GrootType) *Array {
	return &Array{value: value}
}

func (array *Array) GraphQLType() graphql.Type {
	return graphql.NewList(array.value.GraphQLType())
}

func (array *Array) ReflectType() reflect.Type {
	return reflect.SliceOf(array.value.ReflectType())
}
