package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type UnionType struct{}

type Union struct {
	name        string
	description string

	members []*Object
	union   *graphql.Union
	builder *SchemaBuilder

	reflectType reflect.Type
}

func (union *Union) GraphQLType() graphql.Type {
	if union.union != nil {
		return union.union
	}

	types := []*graphql.Object{}
	for _, member := range union.members {
		types = append(types, member.GraphQLType().(*graphql.Object))
	}

	union.union = graphql.NewUnion(graphql.UnionConfig{
		Name:        union.name,
		Description: union.description,
		Types:       types,
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			valueType := reflect.TypeOf(p.Value)
			return union.builder.grootTypes[valueType].GraphQLType().(*graphql.Object)
		},
	})

	return union.union
}

func NewUnion(t reflect.Type, builder *SchemaBuilder) *Union {
	var (
		name            = t.Name()
		embeddedStructs = []reflect.Type{}
		union           = &Union{
			name:        name,
			builder:     builder,
			members:     []*Object{},
			reflectType: t,
		}
	)

	if t.Kind() != reflect.Struct {
		panic("union type must be a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			embeddedStructs = append(embeddedStructs, field.Type)
		} else {
			panic("union type cannot have any field other than embedded structs")
		}
	}

	for _, embeddedStruct := range embeddedStructs {
		if embeddedStruct == reflect.TypeOf(UnionType{}) {
			continue
		}

		if object, ok := builder.grootTypes[embeddedStruct].(*Object); ok {
			union.members = append(union.members, object)
		} else {
			union.members = append(union.members, NewObject(embeddedStruct, builder))
		}
	}

	builder.grootTypes[t] = union
	builder.types[name] = union.GraphQLType()
	return union
}

func (union *Union) resolveValue(p graphql.ResolveTypeParams) reflect.Value {
	for _, member := range union.members {
		name := member.reflectType.Name()
		field := reflect.ValueOf(p.Value).FieldByName(name)

		if !field.IsZero() {
			return field
		}
	}

	panic("could not resolve type")
}
