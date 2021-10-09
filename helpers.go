package groot

import (
	"math/rand"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getOrCreateType(t parser.Type, builder *SchemaBuilder) graphql.Type {
	if graphqlType, ok := builder.getType(t); ok {
		return NewNonNull(graphqlType)
	}

	switch t := t.(type) {
	case *parser.Scalar:
		return NewNonNull(NewScalar(t, builder))
	case *parser.Enum:
		return NewNonNull(NewEnum(t, builder))
	case *parser.Object:
		return NewNonNull(NewObject(t, builder))
	case *parser.Interface:
		return NewNonNull(NewInterface(t, builder))
	case *parser.Union:
		return NewNonNull(NewUnion(t, builder))
	case *parser.Input:
		return NewNonNull(NewInputObject(t, builder))
	case *parser.Array:
		return NewNonNull(NewArray(getOrCreateType(t.Element(), builder)))
	case *parser.Nullable:
		return GetNullable(getOrCreateType(t.Element(), builder)).(graphql.Type)
	}

	panic("groot: unexpected error occurred")
}
