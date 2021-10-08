package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type Array struct {
	value GrootType
}

func NewArray(t *parser.Type, builder *SchemaBuilder) (*Array, error) {
	if t.Kind() != parser.List {
		panic("expected list")
	}

	element := t.Element()
	value, err := getOrCreateType(element, builder)
	if err != nil {
		return nil, err
	}

	return &Array{
		value: value,
	}, nil
}

func (array *Array) GraphQLType() graphql.Type {
	return graphql.NewList(array.value.GraphQLType())
}

func (array *Array) ReflectType() reflect.Type {
	return reflect.SliceOf(array.value.ReflectType())
}

func getOrCreateType(t *parser.Type, builder *SchemaBuilder) (GrootType, error) {
	if grootType, ok := builder.getType(t); ok {
		return NewNonNull(grootType), nil
	}

	var grootType GrootType
	var err error

	switch t.Kind() {
	case parser.Scalar, parser.CustomScalar:
		grootType, err = NewScalar(t, builder)
	case parser.Object:
		grootType, err = NewObject(t, builder)
	case parser.Interface:
		grootType, err = NewInterface(t, builder)
	case parser.InterfaceDefinition:
		grootType, err = NewInterface(t, builder)
	case parser.Union:
		grootType, err = NewUnion(t, builder)
	case parser.Enum:
		grootType, err = NewEnum(t, builder)
	case parser.List:
		grootType, err = NewArray(t, builder)
	case parser.Nullable:
		gType, err := getOrCreateType(t.Element(), builder)
		if err != nil {
			return nil, err
		}

		return GetNullable(gType), nil
	}

	// TODO: default panic

	if err != nil {
		return nil, err
	}

	return NewNonNull(grootType), nil
}
