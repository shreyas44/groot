package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type Field struct {
	Name              string
	Description       string
	DeprecationReason string
	Type              graphql.Type
	Resolve           func(p graphql.ResolveParams) (interface{}, error)
	field             *graphql.Field
}

func (field *Field) GraphQLType() *graphql.Field {
	if field.field != nil {
		return field.field
	}

	field.field = &graphql.Field{
		Name:              field.Name,
		Type:              field.Type,
		Description:       field.Description,
		Resolve:           field.Resolve,
		DeprecationReason: field.DeprecationReason,
	}

	return field.field
}

func test(p graphql.ResolveParams) (interface{}, error) {
	return "world", nil
}

func NewField(structField reflect.StructField) *Field {
	var name string
	var description string
	var depractionReason string

	if ignoreTag := structField.Tag.Get("groot_ignore"); ignoreTag == "true" {
		return nil
	}

	if nameTag := structField.Tag.Get("json"); nameTag != "" {
		name = nameTag
	} else {
		name = structField.Name
	}

	if descTag := structField.Tag.Get("description"); descTag != "" {
		description = descTag
	}

	if deprecate := structField.Tag.Get("deprecate"); deprecate != "" {
		depractionReason = deprecate
	}

	graphqlType, ok := graphqlTypes[structField.Type]
	if structFieldType := structField.Type; !ok && structFieldType.Kind() == reflect.Struct {
		graphqlTypes[structField.Type] = nil
		object := NewObject(structFieldType)
		graphqlType = object.GraphQLType()
	}

	field := &Field{
		Name:              name,
		Description:       description,
		Type:              graphqlType,
		Resolve:           test,
		DeprecationReason: depractionReason,
	}

	// hydrate field.field with *graphql.Field
	field.GraphQLType()

	return field
}

func FieldsFromFields(fields []*Field) []*graphql.Field {
	graphqlFields := []*graphql.Field{}

	for _, field := range fields {
		graphqlField := field.GraphQLType()
		graphqlFields = append(graphqlFields, graphqlField)
	}

	return graphqlFields
}
