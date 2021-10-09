package groot

import (
	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

func NewObject(parserObject *parser.Object, builder *SchemaBuilder) *graphql.Object {
	interfaceCount := len(parserObject.Interfaces())
	interfaces := make([]*graphql.Interface, interfaceCount)
	fields := graphql.Fields{}

	object := graphql.NewObject(graphql.ObjectConfig{
		Name:       parserObject.Name(),
		Interfaces: interfaces,
		Fields:     fields,
	})

	builder.addType(parserObject, object)

	for i, parserInterface := range parserObject.Interfaces() {
		interface_ := getOrCreateType(parserInterface, builder)
		interfaces[i] = GetNullable(interface_).(*graphql.Interface)
	}

	for _, parserField := range parserObject.Fields() {
		object.AddFieldConfig(parserField.JSONName(), NewField(parserField, builder))
	}

	return object
}
