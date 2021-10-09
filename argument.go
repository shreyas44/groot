package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

func NewArgument(parserArg *parser.Argument, builder *SchemaBuilder) *graphql.ArgumentConfig {
	graphqlType := getOrCreateType(parserArg.ArgType(), builder)
	argument := &graphql.ArgumentConfig{
		Type:         graphqlType,
		Description:  parserArg.Description(),
		DefaultValue: parserArg.DefaultValue(),
	}

	return argument
}
