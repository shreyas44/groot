package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

func NewInputObject(input *parser.Input, builder *SchemaBuilder) *graphql.InputObject {
	// TODO: description
	object := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   input.ReflectType().Name(),
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	builder.addType(input, object)
	for _, arg := range input.Arguments() {
		config := &graphql.InputObjectFieldConfig{
			Type:        NewArgument(arg, builder).Type,
			Description: arg.Description(),
		}

		if arg.DefaultValue() != "" {
			config.DefaultValue = arg.DefaultValue()
		}

		object.AddFieldConfig(arg.JSONName(), config)
	}

	return object
}
