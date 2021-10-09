package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
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

func (n *NonNull) ParserType() parser.Type {
	return n.value.ParserType()
}

func GetNullable(t GrootType) GrootType {
	if t, ok := t.(*NonNull); ok {
		return t.value
	}

	return t
}
