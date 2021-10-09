package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type Argument struct {
	name        string
	description string
	type_       GrootType
	default_    string
	argument    *graphql.ArgumentConfig
}

func NewArgument(parserArg *parser.Argument, builder *SchemaBuilder) (*Argument, error) {
	grootType, err := getOrCreateType(parserArg.ArgType(), builder)
	if err != nil {
		return nil, err
	}

	argument := &Argument{
		name:        parserArg.JSONName(),
		description: parserArg.Description(),
		default_:    parserArg.DefaultValue(),
		type_:       grootType,
	}

	return argument, nil
}

func (field *Argument) GraphQLArgument() *graphql.ArgumentConfig {
	if field.argument != nil {
		return field.argument
	}

	field.argument = &graphql.ArgumentConfig{
		Type:         field.type_.GraphQLType(),
		Description:  field.description,
		DefaultValue: field.default_,
	}

	return field.argument
}
