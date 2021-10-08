package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type UnionType = parser.UnionType

type Union struct {
	name        string
	description string

	members []*Object
	union   *graphql.Union
	builder *SchemaBuilder

	reflectType reflect.Type
}

func NewUnion(t *parser.Type, builder *SchemaBuilder) (*Union, error) {
	if t.Kind() != parser.Union {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewUnion must have parser type ParserUnion, received %s",
			t.Name(),
			t.Kind(),
		)
		panic(err)
	}

	var (
		name  = t.Name()
		union = &Union{
			name:        name,
			builder:     builder,
			members:     []*Object{},
			reflectType: t.Type,
		}
	)

	builder.addType(t, union)

	for _, member := range t.Members() {
		obj, err := getOrCreateType(member, builder)
		if err != nil {
			return nil, err
		}

		union.members = append(union.members, GetNullable(obj).(*Object))
	}

	return union, nil
}

func (union *Union) GraphQLType() graphql.Type {
	if union.union != nil {
		return union.union
	}

	placeholderTypes := []*graphql.Object{}
	for range union.members {
		placeholderTypes = append(placeholderTypes, graphql.NewObject(graphql.ObjectConfig{
			Name:   randSeq(10),
			Fields: graphql.Fields{},
		}))
	}

	union.union = graphql.NewUnion(graphql.UnionConfig{
		Name:        union.name,
		Description: union.description,
		Types:       placeholderTypes,
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			// TODO
			valueType := reflect.TypeOf(p.Value)
			return union.builder.reflectGrootMap[valueType].GraphQLType().(*graphql.Object)
		},
	})

	types := union.union.Types()
	for i, member := range union.members {
		// we're changing the underlying value in the slice
		types[i] = member.GraphQLType().(*graphql.Object)
	}

	return union.union
}

func (union *Union) ReflectType() reflect.Type {
	return union.reflectType
}

func (union *Union) resolveValue(p graphql.ResolveTypeParams) reflect.Value {
	for _, member := range union.members {
		name := member.reflectType.Name()
		field := reflect.ValueOf(p.Value).FieldByName(name)

		if !field.IsZero() {
			return field
		}
	}

	firstValue := reflect.ValueOf(p.Value).FieldByName(union.members[0].reflectType.Name())
	return firstValue
}
