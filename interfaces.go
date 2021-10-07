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
		if err := validateInterface(t); err != nil {
			panic(err)
		}

		name = t.Name()
		interfaceDefinition = t.Method(0).Type.Out(0)
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

func validateInterface(t reflect.Type) error {
	if t.NumMethod() != 1 {
		return fmt.Errorf(
			"interface %s can have only one method",
			t.Name(),
		)
	}

	method := t.Method(0)

	if method.Type.NumIn() != 0 {
		return fmt.Errorf(
			"method %s on interface %s should not have input arguments",
			method.Name,
			t.Name(),
		)
	}

	if method.Type.NumOut() != 1 {
		return fmt.Errorf(
			"method %s on interface %s should return exactly one value",
			method.Name,
			t.Name(),
		)
	}

	interfaceDefinition := method.Type.Out(0)

	if parserType, _ := getParserType(interfaceDefinition); parserType != ParserInterfaceDefinition {
		return fmt.Errorf(
			"method %s on interface %s should return a struct with groot.InterfaceType embedded",
			method.Name,
			t.Name(),
		)
	}

	if t.Name()+"Definition" != interfaceDefinition.Name() {
		return fmt.Errorf(
			"method %s on interface %s should return a struct named %sDefinition",
			method.Name,
			t.Name(),
			t.Name(),
		)
	}

	return nil
}
