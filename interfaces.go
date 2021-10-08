package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type InterfaceType = parser.InterfaceType

type Interface struct {
	name        string
	description string
	builder     *SchemaBuilder
	fields      []*Field
	interface_  *graphql.Interface
	reflectType reflect.Type
}

func NewInterface(t *parser.Type, builder *SchemaBuilder) (*Interface, error) {
	if t.Kind() != parser.InterfaceDefinition && t.Kind() != parser.Interface {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewInterface must have parser type ParserInterfaceDefinition or ParserInterface, received %s",
			t.Name(),
			t.Kind(),
		)
		panic(err)
	}

	var interfaceDefinition *parser.Type

	interface_ := &Interface{
		name:        t.Name(),
		builder:     builder,
		reflectType: t.Type,
	}
	builder.addType(t, interface_)

	if t.Kind() == parser.Interface {
		interfaceDefinition = t.Definition()
	} else {
		name := interface_.name
		interface_.name = name[0 : len(name)-len("Definition")]
		interfaceDefinition = t
	}

	fields, err := getFields(interfaceDefinition, builder)
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

func (i *Interface) ReflectType() reflect.Type {
	return i.reflectType
}
