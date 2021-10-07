package groot

import (
	"fmt"
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

func NewUnion(t reflect.Type, builder *SchemaBuilder) *Union {
	if parserType, _ := getParserType(t); parserType != ParserUnion {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewUnion must have parser type ParserUnion, received %s",
			t.Name(),
			parserType,
		)
		panic(err)
	}

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

	builder.grootTypes[t] = union

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

	return union
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
			valueType := reflect.TypeOf(p.Value)
			return union.builder.grootTypes[valueType].GraphQLType().(*graphql.Object)
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

func validateUnionType() {}

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
