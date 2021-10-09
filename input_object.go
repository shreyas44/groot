package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type InputObject struct {
	name        string
	description string
	object      *graphql.InputObject
	fields      []*Argument
	parserInput *parser.Input
}

func NewInputObject(input *parser.Input, builder *SchemaBuilder) (*InputObject, error) {
	inputObject := &InputObject{
		name:        input.Name(),
		parserInput: input,
	}

	args, err := getArguments(input.Arguments(), builder)
	if err != nil {
		return nil, err
	}

	inputObject.fields = args
	builder.addType(input, inputObject)
	return inputObject, nil
}

func (object *InputObject) GraphQLType() graphql.Type {
	if object.object != nil {
		return object.object
	}

	object.object = graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        object.name,
		Fields:      graphql.InputObjectConfigFieldMap{},
		Description: object.description,
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

func (object *InputObject) ParserType() parser.Type {
	return object.parserInput
}
