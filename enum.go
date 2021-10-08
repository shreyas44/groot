package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type EnumType = parser.EnumType

type Enum struct {
	name        string
	values      []string
	enum        *graphql.Enum
	reflectType reflect.Type
}

func NewEnum(t *parser.Type, builder *SchemaBuilder) (*Enum, error) {
	if t.Kind() != parser.Enum {
		err := fmt.Sprintf(
			"groot: reflect.Type %s passed to NewEnum must have parser type of ParserEnum, received %s",
			t.Name(),
			t.Kind(),
		)
		panic(err)
	}

	name := t.Name()
	enumType := reflect.New(t.Type).Interface().(EnumType)
	enum := &Enum{
		name:        name,
		values:      enumType.Values(),
		reflectType: t.Type,
	}

	builder.addType(t, enum)
	return enum, nil
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

	return enum.enum
}

func (enum *Enum) ReflectType() reflect.Type {
	return enum.reflectType
}
