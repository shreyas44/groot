package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type Argument struct {
	name        string
	description string
	type_       graphql.Input
	default_    string
	argument    *graphql.ArgumentConfig
}

func (field *Argument) GraphQLArgument() *graphql.ArgumentConfig {
	if field.argument != nil {
		return field.argument
	}

	field.argument = &graphql.ArgumentConfig{
		Type:         field.type_,
		Description:  field.description,
		DefaultValue: field.default_,
	}

	return field.argument
}

func parseNullableArgument(t reflect.Type, builder *SchemaBuilder) graphql.Input {
	return graphql.GetNullable(parseArgumentType(t.Elem(), builder)).(graphql.Input)
}

func parseArrayArgument(t reflect.Type, builder *SchemaBuilder) graphql.Input {
	return graphql.NewList(parseArgumentType(t.Elem(), builder))
}

func parseObjectArgument(t reflect.Type, builder *SchemaBuilder) graphql.Input {
	if object, ok := builder.types[t.Name()]; ok {
		return object.(graphql.Input)
	}

	if object, ok := builder.grootTypes[t]; ok {
		if object, ok := object.(*InputObject); ok {
			return object.GraphQLType()
		}
	}

	object := NewInputObject(t, builder)
	return object.GraphQLType()
}

func parseScalarArgument(t reflect.Type, builder *SchemaBuilder) graphql.Input {
	scalars := map[reflect.Kind]graphql.Type{
		reflect.Int:     graphql.Int,
		reflect.String:  graphql.String,
		reflect.Bool:    graphql.Boolean,
		reflect.Float32: graphql.Float,
	}

	return scalars[t.Kind()]
}

func parseArgumentType(t reflect.Type, builder *SchemaBuilder) graphql.Input {
	switch t.Kind() {
	case reflect.Ptr:
		return parseNullableArgument(t, builder)
	case reflect.Array:
		return graphql.NewNonNull(parseArrayArgument(t, builder))
	case reflect.Struct:
		return graphql.NewNonNull(parseObjectArgument(t, builder))
	case reflect.Int, reflect.Float32, reflect.String, reflect.Bool:
		return graphql.NewNonNull(parseScalarArgument(t, builder))
	}

	panic("invalid argument type")
}

func NewArgument(structField reflect.StructField, builder *SchemaBuilder) *Argument {
	var (
		name             string
		description      string
		defaultValue     string
		graphqlInputType = parseArgumentType(structField.Type, builder)
	)

	if ignoreTag := structField.Tag.Get("groot_ignore"); ignoreTag == "true" {
		return nil
	}

	if jsonTag := structField.Tag.Get("json"); jsonTag != "" {
		name = jsonTag
	} else {
		name = structField.Name
	}

	if defaultTag := structField.Tag.Get("default"); defaultTag != "" {
		defaultValue = defaultTag
	}

	argument := &Argument{
		name:        name,
		description: description,
		default_:    defaultValue,
		type_:       graphqlInputType,
	}

	return argument
}
