package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type UnionType = parser.UnionType

func NewUnion(parserUnion *parser.Union, builder *SchemaBuilder) *graphql.Union {
	placeholderTypes := []*graphql.Object{}
	for range parserUnion.Members() {
		placeholderTypes = append(placeholderTypes, graphql.NewObject(graphql.ObjectConfig{
			Name:   randSeq(10),
			Fields: graphql.Fields{},
		}))
	}

	// TODO: description
	union := graphql.NewUnion(graphql.UnionConfig{
		Name:  parserUnion.Name(),
		Types: placeholderTypes,
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			valueType := reflect.TypeOf(p.Value)
			return builder.reflectGrootMap[valueType].(*graphql.Object)
		},
	})

	builder.addType(parserUnion, union)

	types := union.Types()
	for i, parserObject := range parserUnion.Members() {
		// we're changing the underlying value in the slice
		types[i] = NewObject(parserObject, builder)
	}

	return union
}

func resolveUnionValue(union *parser.Union, p graphql.ResolveTypeParams) reflect.Value {
	for _, member := range union.Members() {
		name := member.Name()
		field := reflect.ValueOf(p.Value).FieldByName(name)

		if !field.IsZero() {
			return field
		}
	}

	firstValue := reflect.ValueOf(p.Value).FieldByName(union.Members()[0].Name())
	return firstValue
}
