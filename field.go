package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

func NewField(parserField *parser.Field, builder *SchemaBuilder) *graphql.Field {
	var (
		resolver    fieldResolver
		subscribe   fieldSubscriber
		graphqlType = getOrCreateType(parserField.Type(), builder)
	)

	if parserField.Subscriber() != nil {
		// subscription resolver
		subscribe = newFieldSubscriber(parserField.Subscriber(), parserField.Type())
		resolver = func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source, nil
		}

	} else if parserField.Resolver() != nil {
		resolver = newFieldResolver(parserField)
	}

	args := graphql.FieldConfigArgument{}
	for _, parserArgs := range parserField.ArgsInput().Arguments() {
		args[parserArgs.JSONName()] = NewArgument(parserArgs, builder)
	}

	field := &graphql.Field{
		Name:              parserField.JSONName(),
		Type:              graphqlType,
		Description:       parserField.Description(),
		Resolve:           resolver,
		DeprecationReason: parserField.DeprecationReason(),
		Args:              args,
	}

	if subscribe != nil {
		field.Subscribe = subscribe
	}

	return field
}
