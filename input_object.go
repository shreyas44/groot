package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type InputObject struct {
	Name        string
	Description string
	object      *graphql.InputObject
	fields      []*Argument
}

func (object *InputObject) GraphQLType() *graphql.InputObject {
	if object.object != nil {
		for _, field := range object.fields {
			object.object.AddFieldConfig(field.Name, &graphql.InputObjectFieldConfig{
				Type:         field.GraphQLType().Type,
				Description:  field.Description,
				DefaultValue: field.Default,
			})
		}

		return object.object
	}

	fields := graphql.InputObjectConfigFieldMap{}
	for _, field := range object.fields {
		fields[field.Name] = &graphql.InputObjectFieldConfig{
			Type:         field.Type,
			Description:  field.Description,
			DefaultValue: field.Default,
		}
	}

	object.object = graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        object.Name,
		Fields:      fields,
		Description: object.Description,
	})

	return object.object
}

func NewInputObject(t reflect.Type) *InputObject {
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("must pass a reflect type of kind reflect.Struct, received %s", t.Kind()))
	}

	structName := t.Name()
	inputObject := &InputObject{
		Name:   structName,
		fields: []*Argument{},
	}

	graphqlInputTypes[t] = inputObject.GraphQLType()

	structFieldCount := t.NumField()
	for i := 0; i < structFieldCount; i++ {
		structField := t.Field(i)
		field := NewArgument(structField)

		// field is a relationship if it's nil
		if field != nil {
			inputObject.fields = append(inputObject.fields, field)
		}
	}

	inputObject.GraphQLType()
	return inputObject
}
