package groot

import (
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

	parserObject parser.Type
}

func NewObject(parserObject *parser.Object, builder *SchemaBuilder) (*Object, error) {
	var (
		structName = parserObject.Name()
		object     = &Object{
			name:         structName,
			interfaces:   []*Interface{},
			builder:      builder,
			parserObject: parserObject,
		}
	)

	builder.addType(parserObject, object)

	for _, interfaceType := range parserObject.Interfaces() {
		interface_, err := getOrCreateType(interfaceType, builder)
		if err != nil {
			return nil, err
		}

		interface_ = GetNullable(interface_)
		object.interfaces = append(object.interfaces, interface_.(*Interface))
	}

	fields, err := getFields(parserObject, builder)
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

func (object *Object) ParserType() parser.Type {
	return object.parserObject
}

func getFields(t parser.TypeWithFields, builder *SchemaBuilder) ([]*Field, error) {
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
