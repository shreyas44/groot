package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
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

func NewObject(t *parser.Type, builder *SchemaBuilder) (*Object, error) {
	if t.Kind() != parser.Object {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewObject must have parser type ParserObject, received %s",
			t.Name(),
			t.Kind(),
		)
		panic(err)
	}

	var (
		structName = t.Name()
		object     = &Object{
			name:        structName,
			interfaces:  []*Interface{},
			builder:     builder,
			reflectType: t.Type,
		}
	)

	builder.addType(t, object)

	if t.Kind() == parser.Object {
		for _, interfaceType := range t.Interfaces() {
			interface_, err := getOrCreateType(interfaceType, builder)
			if err != nil {
				return nil, err
			}

			interface_ = GetNullable(interface_)
			object.interfaces = append(object.interfaces, interface_.(*Interface))
		}
	}

	fields, err := getFields(t, builder)
	if err != nil {
		return nil, err
	}

	object.fields = fields
	return object, nil
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

func getFields(t *parser.Type, builder *SchemaBuilder) ([]*Field, error) {
	fields := []*Field{}

	for _, parserField := range t.Fields() {
		field, err := NewField(parserField, builder)
		if err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	return fields, nil
}
