package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

func NewInputObject(input *parser.Input, builder *SchemaBuilder) *graphql.InputObject {
	// TODO: description
	object := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   input.Name(),
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	builder.addType(input, object)
	for _, arg := range input.Arguments() {
		object.AddFieldConfig(arg.JSONName(), &graphql.InputObjectFieldConfig{
			Type:         NewArgument(arg, builder).Type,
			Description:  arg.Description(),
			DefaultValue: arg.DefaultValue(),
		})
	}

	return object
}
