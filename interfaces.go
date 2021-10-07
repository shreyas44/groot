package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type InterfaceType struct{}

type Interface struct {
	name        string
	description string
	builder     *SchemaBuilder
	fields      []*Field
	interface_  *graphql.Interface
	reflectType reflect.Type
}

func NewInterface(t reflect.Type, builder *SchemaBuilder) *Interface {
	parserType, _ := getParserType(t)
	if parserType != ParserInterface && parserType != ParserInterfaceDefinition {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewInterface must have parser type ParserInterface or ParserInterfaceDefinition, received %s",
			t.Name(),
			parserType,
		)
		panic(err)
	}

	var (
		interfaceDefinition reflect.Type
		name                string
		interface_          = &Interface{
			builder:     builder,
			reflectType: t,
		}
	)

	builder.grootTypes[t] = interface_
	builder.grootTypes[interfaceDefinition] = interface_

	if parserType == ParserInterfaceDefinition {
		interfaceDefinition = t
		name = t.Name()
		name = name[0 : len(name)-len("Definition")+1]
	} else {
		name = t.Name()

		if t.NumMethod() != 1 {
			panic("interface type can only have one method")
		}

		method := t.Method(0).Type

		if method.NumIn() != 0 {
			panic("interface type method must have no input arguments")
		}

		if method.NumOut() != 1 {
			panic("interface type method must have one output argument")
		}

		interfaceDefinition = method.Out(0)

		if interfaceDefinition.Kind() != reflect.Struct {
			panic("interface type method must return a struct")
		}

		if !isInterfaceDefinition(interfaceDefinition) {
			panic("interface type method must return a struct with groot.InterfaceType embedded")
		}
	}

	interface_.name = name
	interface_.fields = getFields(interfaceDefinition, builder)
	return interface_
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
			return i.builder.grootTypes[valueType].GraphQLType().(*graphql.Object)
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

func validateInterface()           {}
func validateInterfaceDefinition() {}
