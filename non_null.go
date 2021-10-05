package groot

import "github.com/graphql-go/graphql"

type NonNull struct {
	value GrootType
}

func (n *NonNull) GraphQLType() graphql.Type {
	return graphql.NewNonNull(n.value.GraphQLType())
}

func NewNonNull(value GrootType) *NonNull {
	return &NonNull{value: value}
}

func GetNullable(t GrootType) GrootType {
	if t, ok := t.(*NonNull); ok {
		return t.value
	}

	return t
}
