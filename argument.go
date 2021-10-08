package groot

import (
	"fmt"

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

func NewArgument(field *parser.ObjectField, builder *SchemaBuilder) (*Argument, error) {
	object := field.Object()
	if object.Kind() != parser.Object {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewArgument must have parser type ParserObject, received %s",
			object.Name(),
			object.Kind(),
		)
		panic(err)
	}

	if err := validateArgumentType(field); err != nil {
		return nil, err
	}

	grootType, err := getArgumentGrootType(field.Type(), builder)
	if err != nil {
		return nil, err
	}

	argument := &Argument{
		name:        field.JSONName(),
		description: field.Description(),
		default_:    field.DefaultValue(),
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

func validateArgumentType(field *parser.ObjectField) error {
	switch field.Type().Kind() {
	case parser.Interface, parser.Union, parser.InterfaceDefinition:
		return fmt.Errorf(
			"argument type %s not supported for field %s on struct %s \nif you think this is a mistake please open an issue at github.com/shreyas44/groot",
			field.StructField.Type.Name(),
			field.Name,
			field.Object().Name(),
		)
	}

	return nil
}

func getArgumentGrootType(t *parser.Type, builder *SchemaBuilder) (GrootType, error) {
	if grootType, ok := builder.getType(t); ok {
		return NewNonNull(grootType), nil
	}

	var argType GrootType
	var err error

	switch t.Kind() {
	case parser.Scalar, parser.CustomScalar:
		argType, err = NewScalar(t, builder)
	case parser.Object:
		argType, err = NewInputObject(t, builder)
	case parser.List:
		argType, err = NewArray(t, builder)
	case parser.Nullable:
		itemType, err := getArgumentGrootType(t.Element(), builder)
		if err != nil {
			return nil, err
		}

		return GetNullable(itemType), nil
	case parser.Enum:
		argType, err = NewEnum(t, builder)
	}

	if err != nil {
		return nil, err
	}

	return NewNonNull(argType), nil
}
