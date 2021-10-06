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

func (i *Interface) GraphQLType() graphql.Type {
	if i.interface_ != nil {
		for _, field := range i.fields {
			i.interface_.AddFieldConfig(field.name, field.GraphQLField())
		}

		return i.interface_
	}

	fields := graphql.Fields{}
	for _, field := range i.fields {
		fields[field.name] = field.GraphQLField()
	}

	i.interface_ = graphql.NewInterface(graphql.InterfaceConfig{
		Name:        i.name,
		Description: i.description,
		Fields:      fields,
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			valueType := reflect.TypeOf(p.Value)
			return i.builder.grootTypes[valueType].GraphQLType().(*graphql.Object)
		},
	})

	return i.interface_
}

func isInterfaceDefinition(t reflect.Type) bool {
	interfaceType := reflect.TypeOf(InterfaceType{})

	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		if field := t.Field(i); field.Anonymous && field.Type == interfaceType {
			return true
		}
	}

	return false
}

func NewInterface(t reflect.Type, builder *SchemaBuilder) *Interface {
	if t.Kind() != reflect.Interface && t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("must pass a reflect type of kind reflect.Interface or reflect.Struct, received %s", t.Kind()))
	}

	var (
		interfaceDefinition reflect.Type
		name                string
		interface_          = &Interface{
			builder:     builder,
			reflectType: t,
		}
	)

	if t.Kind() == reflect.Struct {
		interfaceDefinition = t
		name = t.Name()
		name = name[0 : len(name)-len("Definition")+1]
	} else if t.Kind() == reflect.Interface {
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
	builder.grootTypes[t] = interface_
	builder.grootTypes[interfaceDefinition] = interface_
	builder.types[interface_.name] = interface_.GraphQLType()
	interface_.fields = getFields(interfaceDefinition, builder)
	interface_.GraphQLType()
	return interface_
}
