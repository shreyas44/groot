package groot

import (
	"context"
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

// TODO: add better typing support
type InterfaceType interface {
	ResolveInterfaceType(value interface{}, ctx context.Context, info graphql.ResolveInfo) string
}

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
		// TODO: fix value type
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			name := p.Value.(InterfaceType).ResolveInterfaceType(p.Value, p.Context, p.Info)
			return i.builder.types[name].(*graphql.Object)
		},
	})

	return i.interface_
}

func NewInterface(t reflect.Type, builder *SchemaBuilder) *Interface {
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("must pass a reflect type of kind reflect.Struct, received %s", t.Kind()))
	}

	var (
		interfaceType = reflect.TypeOf((*InterfaceType)(nil)).Elem()
		structName    = t.Name()
		interface_    = &Interface{
			name:        structName,
			builder:     builder,
			reflectType: t,
		}
	)

	if !t.Implements(interfaceType) {
		panic(fmt.Sprintf("%s must implement groot.InterfaceType", structName))
	}

	builder.grootTypes[t] = interface_
	builder.types[interface_.name] = interface_.GraphQLType()
	interface_.fields = getFields(t, builder)
	interface_.GraphQLType()
	return interface_
}
