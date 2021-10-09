package groot

import (
	"github.com/graphql-go/graphql"
)

func NewNonNull(t graphql.Type) *graphql.NonNull {
	if _, ok := t.(*graphql.NonNull); ok {
		return t.(*graphql.NonNull)
	}

	return graphql.NewNonNull(t)
}

func GetNullable(t graphql.Type) graphql.Nullable {
	return graphql.GetNullable(t)
}
