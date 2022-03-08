package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type EnumType = parser.EnumType

func NewEnum(t *parser.Enum, builder *SchemaBuilder) *graphql.Enum {
	name := t.ReflectType().Name()
	enumType := reflect.New(t.ReflectType()).Interface().(EnumType)

	values := graphql.EnumValueConfigMap{}
	for _, value := range enumType.Values() {
		values[value] = &graphql.EnumValueConfig{
			Value: value,
		}
	}

	// TODO: enum description, value descriptions, value deprecation
	enum := graphql.NewEnum(graphql.EnumConfig{
		Name:   name,
		Values: values,
	})

	builder.addType(t, enum)
	return enum
}
