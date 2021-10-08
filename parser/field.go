package parser

import (
	"fmt"
	"reflect"
)

type ObjectField struct {
	reflect.StructField
	fieldType         *Type
	object            *Type
	arguments         *Type
	resolver          *Resolver
	subscriber        *Subscriber
	jsonName          string
	description       string
	deprecationReason string
	defaultValue      string
}

func NewObjectField(t *Type, field reflect.StructField) (*ObjectField, error) {
	if field.Tag.Get("groot_ignore") == "true" {
		return nil, nil
	}

	if t.Kind() != Object && t.Kind() != InterfaceDefinition {
		panic("something happened")
	}

	var (
		subscriber  *Subscriber
		resolver    *Resolver
		arguments   *Type
		fieldType   *Type
		err         error
		objectField = &ObjectField{
			StructField:       field,
			object:            t,
			description:       field.Tag.Get("description"),
			jsonName:          field.Tag.Get("json"),
			deprecationReason: field.Tag.Get("deprecate"),
			defaultValue:      field.Tag.Get("default"),
		}
	)

	fieldType, err = getOrCreateType(field.Type)
	if err != nil {
		return nil, err
	}

	objectField.fieldType = fieldType
	if err := validateFieldType(objectField); err != nil {
		return nil, err
	}

	if t.Name() == "Subscription" {
		if subscriber, err = NewResolver(t, objectField); err != nil {
			return nil, err
		}

		if arguments, err = getArguments(subscriber); err != nil {
			return nil, err
		}
	} else {
		if resolver, err = NewResolver(t, objectField); err != nil {
			return nil, err
		}

		if arguments, err = getArguments(resolver); err != nil {
			return nil, err
		}
	}

	objectField.resolver = resolver
	objectField.subscriber = subscriber
	objectField.arguments = arguments
	objectField.fieldType = fieldType
	return objectField, nil
}

func (f *ObjectField) Object() *Type {
	return f.object
}

func (f *ObjectField) Arguments() *Type {
	return f.arguments
}

func (f *ObjectField) Resolver() *Resolver {
	return f.resolver
}

func (f *ObjectField) Subscriber() *Subscriber {
	return f.subscriber
}

func (f *ObjectField) Type() *Type {
	return f.fieldType
}

func (f *ObjectField) Description() string {
	return f.description
}

func (f *ObjectField) JSONName() string {
	if f.jsonName == "" {
		return f.Name
	}

	return f.jsonName
}

func (f *ObjectField) DeprecationReason() string {
	return f.deprecationReason
}

func (f *ObjectField) DefaultValue() string {
	return f.defaultValue
}

func validateFieldType(field *ObjectField) error {
	parserType, err := getTypeKind(field.Type().Type)
	if err != nil {
		return fmt.Errorf(
			"field type %s not supported for field %s on struct %s \nif you think this is a mistake please open an issue at github.com/shreyas44/groot",
			field.Type().Name(),
			field.Name,
			field.Object().Name(),
		)
	}

	if parserType == InterfaceDefinition {
		return fmt.Errorf(
			"received an interface definition for field type %s for field %s on struct %s\n"+
				"create a Go interface corresponding to the GraphQL interface and use that instead\n"+
				"see https://groot.shreyas44.com/type-definitions/interface for more info",
			field.Type().Name(),
			field.StructField.Name,
			field.Object().Name(),
		)
	}

	return nil
}
