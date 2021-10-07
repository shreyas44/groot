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

func NewInputObject(t reflect.Type, builder *SchemaBuilder) *InputObject {
	if parserType, _ := getParserType(t); parserType != ParserObject {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewInputObject must have parser type ParserObject, received %s",
			t.Name(),
			parserType,
		)
		panic(err)
	}

	structName := t.Name()
	inputObject := &InputObject{
		name:        structName,
		fields:      []*Argument{},
		reflectType: t,
	}

	structFieldCount := t.NumField()

	for i := 0; i < structFieldCount; i++ {
		structField := t.Field(i)
		field := NewArgument(t, structField, builder)

		if field != nil {
			inputObject.fields = append(inputObject.fields, field)
		}
	}

	builder.grootTypes[t] = inputObject
	return inputObject
}

func (object *InputObject) GraphQLType() graphql.Type {
	if object.object != nil {
		return object.object
	}

	object.object = graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        object.name,
		Fields:      graphql.InputObjectConfigFieldMap{},
		Description: object.Description,
	})

	for _, field := range object.fields {
		object.object.AddFieldConfig(field.name, &graphql.InputObjectFieldConfig{
			Type:         field.type_.GraphQLType(),
			Description:  field.description,
			DefaultValue: field.default_,
		})
	}

	return object.object
}

func (object *InputObject) ReflectType() reflect.Type {
	return object.reflectType
}
