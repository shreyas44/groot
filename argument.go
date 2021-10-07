package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type Argument struct {
	name        string
	description string
	type_       GrootType
	default_    string
	argument    *graphql.ArgumentConfig
}

func NewArgument(structType reflect.Type, structField reflect.StructField, builder *SchemaBuilder) *Argument {
	if parserType, _ := getParserType(structType); parserType != ParserObject {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewArgument must have parser type ParserObject, received %s",
			structType.Name(),
			parserType,
		)
		panic(err)
	}

	var (
		name           string
		description    string
		defaultValue   string
		grootType, err = getArgumentGrootType(structType, structField, builder)
	)

	if err != nil {
		panic(err)
	}

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
		type_:       grootType,
	}

	return argument
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

func validateArgumentType(structType reflect.Type, structField reflect.StructField) error {
	notSupportError := fmt.Errorf(
		"argument type %s not supported for field %s on struct %s \nif you think this is a mistake please open an issue at github.com/shreyas44/groot",
		structField.Type.Name(),
		structField.Name,
		structType.Name(),
	)

	parserType, err := getParserType(structField.Type)
	if err != nil || parserType == ParserInterface || parserType == ParserUnion || parserType == ParserInterfaceDefinition {
		return notSupportError
	}

	return nil
}

func getArgumentGrootType(structType reflect.Type, structField reflect.StructField, builder *SchemaBuilder) (GrootType, error) {
	if parserType, _ := getParserType(structType); parserType != ParserObject && parserType != ParserInterfaceDefinition {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to getArgumentGrootType must have parser type ParserObject or ParserInterfaceDefinition, received %s",
			structType.Name(),
			parserType,
		)
		panic(err)
	}

	parserType, _ := getParserType(structField.Type)
	err := validateArgumentType(structType, structField)
	if err != nil {
		return nil, err
	}

	if grootType := builder.getType(structField.Type); grootType != nil {
		return NewNonNull(grootType), nil
	}

	switch parserType {
	case ParserScalar, ParserCustomScalar:
		return NewNonNull(NewScalar(structField.Type, builder)), nil
	case ParserObject:
		return NewNonNull(NewInputObject(structField.Type, builder)), nil
	case ParserList:
		field := structField
		field.Type = field.Type.Elem()
		itemType, err := getArgumentGrootType(structType, field, builder)
		if err != nil {
			return nil, err
		}

		return NewNonNull(NewArray(itemType)), nil

	case ParserNullable:
		field := structField
		field.Type = field.Type.Elem()
		itemType, err := getArgumentGrootType(structType, field, builder)
		if err != nil {
			return nil, err
		}

		return GetNullable(itemType), nil

	case ParserEnum:
		return NewNonNull(NewEnum(structField.Type, builder)), nil
	}

	// should be unreachable
	panic("groot: invalid argument type")
}
