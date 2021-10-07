package groot

import (
	"fmt"
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

func NewEnum(t reflect.Type, builder *SchemaBuilder) *Enum {
	if parserType, _ := getParserType(t); parserType != ParserEnum {
		err := fmt.Sprintf(
			"groot: reflect.Type %s passed to NewEnum must have parser type of ParserEnum, received %s",
			t.Name(),
			parserType,
		)
		panic(err)
	}

	name := t.Name()
	enumType := reflect.New(t).Interface().(EnumType)
	enum := &Enum{
		name:        name,
		values:      enumType.Values(),
		reflectType: t,
	}

	builder.grootTypes[t] = enum
	return enum
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
