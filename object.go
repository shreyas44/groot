package gql

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

var types = map[string]graphql.Type{
	"string": graphql.String,
	"int": graphql.Int,
	"float32": graphql.Float,
	"bool": graphql.Boolean,
}

func NewObject(t reflect.Type) *graphql.Object {
	objectFields := graphql.Fields{}
	fieldsCount := t.NumField()

	for i := 0; i < fieldsCount; i++ {
		field := t.Field(i)
		fieldType := field.Type.Name()

		gType := types[fieldType]
		gField := graphql.Field{ Type: gType }

		objectFields[field.Name] = &gField
	}

	gObjectConfig := graphql.ObjectConfig{ Fields: objectFields }
	gobject := graphql.NewObject(gObjectConfig)

	types[t.Name()] = gobject

	return gobject
}