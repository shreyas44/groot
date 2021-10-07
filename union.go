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

func NewUnion(t reflect.Type, builder *SchemaBuilder) (*Union, error) {
	if parserType, _ := getParserType(t); parserType != ParserUnion {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewUnion must have parser type ParserUnion, received %s",
			t.Name(),
			parserType,
		)
		panic(err)
	}

	var (
		name  = t.Name()
		union = &Union{
			name:        name,
			builder:     builder,
			members:     []*Object{},
			reflectType: t,
		}
	)

	builder.grootTypes[t] = union

	if err := validateUnionType(t); err != nil {
		return nil, err
	}

	for i := 0; i < t.NumField(); i++ {
		embeddedStruct := t.Field(i).Type

		if embeddedStruct == reflect.TypeOf(UnionType{}) {
			continue
		}

		if object, ok := builder.getType(embeddedStruct).(*Object); ok {
			union.members = append(union.members, object)
		} else {
			object, err := NewObject(embeddedStruct, builder)
			if err != nil {
				return nil, err
			}

			union.members = append(union.members, object)
		}
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

func validateUnionType(t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		parserType, err := getParserType(field.Type)
		if err != nil {
			return err
		}

		if parserType != ParserObject && !field.Anonymous {
			err := fmt.Errorf(
				"got extra field %s on union %s, union types cannot contain any field other than embedded structs and groot.UnionType",
				field.Name,
				t.Name(),
			)
			panic(err)
		}
	}

	return nil
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
