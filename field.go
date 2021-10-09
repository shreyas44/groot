package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

func NewField(parserField *parser.Field, builder *SchemaBuilder) *graphql.Field {
	var resolver graphql.FieldResolveFn
	var subscribe graphql.FieldResolveFn

	graphqlType := getOrCreateType(parserField.Type(), builder)

	// default resolver
	resolver = func(p graphql.ResolveParams) (interface{}, error) {
		value := reflect.ValueOf(p.Source).FieldByName(parserField.Name)
		return value.Interface(), nil
	}

	if parserField.Subscriber() != nil {
		// subscription resolver
		subscribe = buildSubscriptionResolver(parserField.Subscriber(), parserField.Type())
		resolver = func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source, nil
		}

	} else if parserField.Resolver() != nil {
		// custom resolver
		resolver = buildResolver(parserField.Resolver(), parserField.Type())
	}

	args := graphql.FieldConfigArgument{}
	for _, parserArgs := range parserField.Arguments() {
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
