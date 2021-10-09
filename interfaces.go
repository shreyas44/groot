package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type InterfaceType = parser.InterfaceType

func NewInterface(parserInterface *parser.Interface, builder *SchemaBuilder) *graphql.Interface {
	// TODO: description
	interface_ := graphql.NewInterface(graphql.InterfaceConfig{
		Name:   parserInterface.Name(),
		Fields: graphql.Fields{},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			valueType := reflect.TypeOf(p.Value)
			return builder.reflectGrootMap[valueType].(*graphql.Object)
		},
	})

	builder.addType(parserInterface, interface_)
	for _, field := range parserInterface.Fields() {
		interface_.AddFieldConfig(field.JSONName(), NewField(field, builder))
	}

	return interface_
}
