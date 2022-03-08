package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

func NewField(parserField *parser.Field, builder *SchemaBuilder) *graphql.Field {
	var (
		subscribe   fieldSubscriber
		graphqlType = getOrCreateType(parserField.Type(), builder)
	)

	if parserField.Subscriber() != nil {
		subscribe = newFieldSubscriber(parserField.Subscriber(), parserField.Type())
	}

	args := graphql.FieldConfigArgument{}
	for _, parserArgs := range parserField.ArgsInput().Arguments() {
		args[parserArgs.JSONName()] = NewArgument(parserArgs, builder)
	}

	field := &graphql.Field{
		Name:              parserField.JSONName(),
		Type:              graphqlType,
		Description:       parserField.Description(),
		Resolve:           newFieldResolver(parserField),
		DeprecationReason: parserField.DeprecationReason(),
		Args:              args,
		Subscribe:         subscribe,
	}

	return field
}
