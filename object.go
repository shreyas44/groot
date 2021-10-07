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

func NewObject(t reflect.Type, builder *SchemaBuilder) *Object {
	if parserType, _ := getParserType(t); parserType != ParserObject {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewObject must have parser type ParserObject, received %s",
			t.Name(),
			parserType,
		)
		panic(err)
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

	builder.grootTypes[t] = object

	for i := 0; i < structFieldCount; i++ {
		if field := t.Field(i); field.Anonymous && isInterfaceDefinition(field.Type) {
			if interface_, ok := builder.grootTypes[field.Type].(*Interface); ok {
				object.interfaces = append(object.interfaces, interface_)
			} else {
				object.interfaces = append(object.interfaces, NewInterface(field.Type, builder))
			}
		}
	}

	object.fields = getFields(t, builder)
	return object
}

func (object *Object) GraphQLType() graphql.Type {
	if object.object != nil {
		return object.object
	}

	interfaces := []*graphql.Interface{}
	for _, interface_ := range object.interfaces {
		interfaces = append(interfaces, interface_.GraphQLType().(*graphql.Interface))
	}

	object.object = graphql.NewObject(graphql.ObjectConfig{
		Name:        object.name,
		Description: object.description,
		Fields:      graphql.Fields{},
		Interfaces:  interfaces,
	})

	for _, field := range object.fields {
		object.object.AddFieldConfig(field.name, field.GraphQLField())
	}

	return object.object
}

func (object *Object) ReflectType() reflect.Type {
	return object.reflectType
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
