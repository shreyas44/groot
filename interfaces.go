package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type InterfaceType = parser.InterfaceType

type Interface struct {
	name            string
	description     string
	builder         *SchemaBuilder
	fields          []*Field
	interface_      *graphql.Interface
	parserInterface *parser.Interface
}

func NewInterface(parserInterface *parser.Interface, builder *SchemaBuilder) (*Interface, error) {
	interface_ := &Interface{
		name:            parserInterface.Name(),
		builder:         builder,
		parserInterface: parserInterface,
	}

	builder.addType(parserInterface, interface_)

	fields, err := getFields(parserInterface, builder)
	if err != nil {
		return nil, err
	}

	interface_.fields = fields
	return interface_, nil
}

func (i *Interface) GraphQLType() graphql.Type {
	if i.interface_ != nil {
		return i.interface_
	}

	i.interface_ = graphql.NewInterface(graphql.InterfaceConfig{
		Name:        i.name,
		Description: i.description,
		Fields:      graphql.Fields{},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			valueType := reflect.TypeOf(p.Value)
			return i.builder.reflectGrootMap[valueType].GraphQLType().(*graphql.Object)
		},
	})

	for _, field := range i.fields {
		i.interface_.AddFieldConfig(field.name, field.GraphQLField())
	}

	return i.interface_
}

func (i *Interface) ParserType() parser.Type {
	return i.parserInterface
}
