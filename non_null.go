package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type NonNull struct {
	value GrootType
}

func NewNonNull(value GrootType) *NonNull {
	if _, ok := value.(*NonNull); ok {
		return value.(*NonNull)
	}

	return &NonNull{value: value}
}

func (n *NonNull) GraphQLType() graphql.Type {
	return graphql.NewNonNull(n.value.GraphQLType())
}

func (n *NonNull) ReflectType() reflect.Type {
	return n.value.ReflectType()
}

func GetNullable(t GrootType) GrootType {
	if t, ok := t.(*NonNull); ok {
		return t.value
	}

	return t
}
