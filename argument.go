package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type Argument struct {
	Name        string
	Description string
	Type        graphql.Input
	Default     string // avoid interface{}
	argument    *graphql.ArgumentConfig
}

func (field *Argument) GraphQLType() *graphql.ArgumentConfig {
	if field.argument != nil {
		return field.argument
	}

	field.argument = &graphql.ArgumentConfig{
		Type:         field.Type,
		Description:  field.Description,
		DefaultValue: field.Default,
	}

	return field.argument
}

func NewArgument(structField reflect.StructField) *Argument {
	var name string
	var description string
	var defaultValue string

	if ignoreTag := structField.Tag.Get("groot_ignore"); ignoreTag == "true" {
		return nil
	}

	if jsonTag := structField.Tag.Get("json"); jsonTag != "" {
		name = jsonTag
	} else {
		name = structField.Name
	}

	if defaultTag := structField.Tag.Get("default"); defaultTag != "" {
		defaultValue = defaultTag
	}

	graphqlInputType, ok := graphqlInputTypes[structField.Type]
	if structFieldType := structField.Type; !ok && structFieldType.Kind() == reflect.Struct {
		graphqlTypes[structField.Type] = nil
		inputObject := NewInputObject(structFieldType)
		graphqlInputType = inputObject.GraphQLType()
	}

	argument := &Argument{
		Name:        name,
		Description: description,
		Default:     defaultValue,
		Type:        graphqlInputType,
	}

	return argument
}
