package groot

import (
	"github.com/graphql-go/graphql"
)

func NewArray(t graphql.Type) graphql.Type {
	return graphql.NewList(t)
}
