package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type EnumType = parser.EnumType

type Enum struct {
	name       string
	values     []string
	enum       *graphql.Enum
	parserEnum *parser.Enum
}

func NewEnum(t *parser.Enum, builder *SchemaBuilder) (*Enum, error) {
	name := t.Name()
	enumType := reflect.New(t.Type).Interface().(EnumType)
	enum := &Enum{
		name:       name,
		values:     enumType.Values(),
		parserEnum: t,
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

func (enum *Enum) ParserType() parser.Type {
	return enum.parserEnum
}
