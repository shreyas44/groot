package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type Object struct {
	name        string
	description string
	object      *graphql.Object
	fields      []*Field

	builder    *SchemaBuilder
	interfaces []*Interface

	reflectType reflect.Type
}

func (object *Object) GraphQLType() graphql.Type {
	if object.object != nil {
		for _, field := range object.fields {
			object.object.AddFieldConfig(field.name, field.GraphQLField())
		}

		return object.object
	}

	fields := graphql.Fields{}
	for _, field := range object.fields {
		fields[field.name] = field.GraphQLField()
	}

	interfaces := []*graphql.Interface{}
	for _, interface_ := range object.interfaces {
		interfaces = append(interfaces, interface_.GraphQLType().(*graphql.Interface))
	}

	object.object = graphql.NewObject(graphql.ObjectConfig{
		Name:        object.name,
		Description: object.description,
		Fields:      fields,
		Interfaces:  interfaces,
	})

	return object.object
}

func getFields(t reflect.Type, builder *SchemaBuilder) []*Field {
	fields := []*Field{}
	fieldCount := t.NumField()

	for i := 0; i < fieldCount; i++ {
		structField := t.Field(i)

		if structField.Anonymous {
			fields = append(fields, getFields(structField.Type, builder)...)
		} else if field := NewField(t, structField, builder); field != nil {
			fields = append(fields, field)
		}
	}

	return fields
}

func NewObject(t reflect.Type, builder *SchemaBuilder) *Object {
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("must pass a reflect type of kind reflect.Struct, received %s", t.Kind()))
	}

	var (
		structName       = t.Name()
		structFieldCount = t.NumField()
		object           = &Object{
			name:        structName,
			interfaces:  []*Interface{},
			builder:     builder,
			reflectType: t,
		}
	)

	for i := 0; i < structFieldCount; i++ {
		if field := t.Field(i); field.Anonymous && isInterfaceDefinition(field.Type) {
			if interface_, ok := builder.grootTypes[field.Type].(*Interface); ok {
				object.interfaces = append(object.interfaces, interface_)
			} else {
				object.interfaces = append(object.interfaces, NewInterface(field.Type, builder))
			}
		}
	}

	builder.grootTypes[t] = object
	builder.types[object.name] = object.GraphQLType()
	object.fields = getFields(t, builder)
	object.GraphQLType()
	return object
}
