package parser

import (
	"fmt"
	"reflect"
)

type Field struct {
	structField       reflect.StructField
	type_             Type
	object            TypeWithFields
	argsInput         *Input
	resolver          *Resolver
	subscriber        *Subscriber
	jsonName          string
	description       string
	deprecationReason string
}

func NewField(t TypeWithFields, field reflect.StructField) (*Field, error) {
	if field.Tag.Get("json") == "-" || !field.IsExported() {
		return nil, nil
	}

	var (
		subscriber  *Subscriber
		resolver    *Resolver
		argsInput   *Input
		fieldType   Type
		err         error
		objectField = &Field{
			structField:       field,
			object:            t,
			description:       field.Tag.Get("description"),
			jsonName:          field.Tag.Get("json"),
			deprecationReason: field.Tag.Get("deprecate"),
		}
	)

	fieldType, err = getOrCreateType(field.Type)
	if err != nil {
		return nil, err
	}

	objectField.type_ = fieldType
	if err := validateFieldType(t.ReflectType(), field); err != nil {
		return nil, err
	}

	if t.ReflectType().Name() == "Subscription" {
		if subscriber, err = NewResolver(objectField); err != nil {
			return nil, err
		}

		if argsInput, err = getResolverArgsInput(subscriber); err != nil {
			return nil, err
		}
	} else {
		if resolver, err = NewResolver(objectField); err != nil {
			return nil, err
		}

		if resolver != nil {
			argsInput, err = getResolverArgsInput(resolver)
			if err != nil {
				return nil, err
			}
		}
	}

	objectField.resolver = resolver
	objectField.subscriber = subscriber
	objectField.argsInput = argsInput
	objectField.type_ = fieldType
	return objectField, nil
}

func (f *Field) Object() TypeWithFields {
	return f.object
}

func (f *Field) ArgsInput() *Input {
	return f.argsInput
}

func (f *Field) Resolver() *Resolver {
	return f.resolver
}

func (f *Field) Subscriber() *Subscriber {
	return f.subscriber
}

func (f *Field) Type() Type {
	return f.type_
}

func (f *Field) Description() string {
	return f.description
}

func (f *Field) JSONName() string {
	if f.jsonName == "" {
		return f.structField.Name
	}

	return f.jsonName
}

func (f *Field) StructField() reflect.StructField {
	return f.structField
}

func (f *Field) DeprecationReason() string {
	return f.deprecationReason
}

func validateFieldType(structType reflect.Type, field reflect.StructField) error {
	parserType, err := getTypeKind(field.Type)
	if err != nil {
		return fmt.Errorf(
			"field type %s not supported for field %s on struct %s \nif you think this is a mistake please open an issue at github.com/shreyas44/groot",
			field.Type.Name(),
			field.Name,
			structType.Name(),
		)
	}

	if parserType == KindInterfaceDefinition {
		return fmt.Errorf(
			"received an interface definition for field type %s for field %s on struct %s\n"+
				"create a Go interface corresponding to the GraphQL interface and use that instead\n"+
				"see https://groot.shreyas44.com/type-definitions/interface for more info",
			field.Type.Name(),
			field.Name,
			structType.Name(),
		)
	}

	return nil
}
