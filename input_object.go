package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type InputObject struct {
	name        string
	Description string
	object      *graphql.InputObject
	fields      []*Argument
	reflectType reflect.Type
}

func (object *InputObject) GraphQLType() graphql.Type {
	if object.object != nil {
		for _, field := range object.fields {
			object.object.AddFieldConfig(field.name, &graphql.InputObjectFieldConfig{
				Type:         field.GraphQLArgument().Type,
				Description:  field.description,
				DefaultValue: field.default_,
			})
		}

		return object.object
	}

	fields := graphql.InputObjectConfigFieldMap{}
	for _, field := range object.fields {
		fields[field.name] = &graphql.InputObjectFieldConfig{
			Type:         field.type_,
			Description:  field.description,
			DefaultValue: field.default_,
		}
	}

	object.object = graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        object.name,
		Fields:      fields,
		Description: object.Description,
	})

	return object.object
}

func NewInputObject(t reflect.Type, builder *SchemaBuilder) *InputObject {
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("must pass a reflect type of kind reflect.Struct, received %s", t.Kind()))
	}

	structName := t.Name()
	inputObject := &InputObject{
		name:        structName,
		fields:      []*Argument{},
		reflectType: t,
	}

	builder.types[structName] = inputObject.GraphQLType()
	structFieldCount := t.NumField()

	for i := 0; i < structFieldCount; i++ {
		structField := t.Field(i)
		field := NewArgument(structField, builder)

		if field != nil {
			inputObject.fields = append(inputObject.fields, field)
		}
	}

	builder.grootTypes[t] = inputObject
	builder.types[structName] = inputObject.GraphQLType()
	return inputObject
}
