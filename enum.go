package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type EnumType interface {
	Values() []string
}

type Enum struct {
	name        string
	values      []string
	enum        *graphql.Enum
	reflectType reflect.Type
}

func (enum *Enum) GraphQLType() graphql.Type {
	if enum.enum != nil {
		return enum.enum
	}

	values := graphql.EnumValueConfigMap{}
	for _, value := range enum.values {
		values[value] = &graphql.EnumValueConfig{
			Value: value,
		}
	}

	// TODO: enum description, value descriptions, value deprecation
	enum.enum = graphql.NewEnum(graphql.EnumConfig{
		Name:   enum.name,
		Values: values,
	})

	graphql.NewEnum(graphql.EnumConfig{
		Name: "UserType",
		Values: graphql.EnumValueConfigMap{
			"ADMIN": &graphql.EnumValueConfig{
				Value: "ADMIN",
			},
			"USER": &graphql.EnumValueConfig{
				Value: "USER",
			},
		},
	})

	return enum.enum
}

func NewEnum(t reflect.Type, builder *SchemaBuilder) *Enum {
	enumInterfaceType := reflect.TypeOf((*EnumType)(nil)).Elem()
	name := t.Name()

	if !t.Implements(enumInterfaceType) {
		panic("enum must implement the groot.EnumType interface")
	}

	enumType := reflect.New(t).Interface().(EnumType)
	enum := &Enum{
		name:        name,
		values:      enumType.Values(),
		reflectType: t,
	}

	builder.grootTypes[t] = enum
	builder.types[name] = enum.GraphQLType()
	return enum
}
