package groot

import (
	"reflect"

	"github.com/shreyas44/groot/parser"
)

func ParseObject(i interface{}) (*parser.Object, error) {
	return parser.ParseObject(reflect.TypeOf(i))
}

func ParseInputObject(i interface{}) (*parser.Input, error) {
	return parser.ParseInputObject(reflect.TypeOf(i))
}

func ParseUnion(i interface{}) (*parser.Union, error) {
	return parser.ParseUnion(reflect.TypeOf(i))
}

func ParseInterface(i interface{}) (*parser.Interface, error) {
	return parser.ParseInterface(reflect.TypeOf(i))
}

func ParseEnum(i interface{}) (*parser.Enum, error) {
	return parser.ParseEnum(reflect.TypeOf(i))
}

func ParseScalar(i interface{}) (*parser.Scalar, error) {
	return parser.ParseScalar(reflect.TypeOf(i))
}

func MustParseObject(i interface{}) *parser.Object {
	object, err := ParseObject(i)
	if err != nil {
		panic(err)
	}

	return object
}

func MustParseInputObject(i interface{}) *parser.Input {
	inputObject, err := ParseInputObject(i)
	if err != nil {
		panic(err)
	}

	return inputObject
}

func MustParseUnion(i interface{}) *parser.Union {
	union, err := ParseUnion(i)
	if err != nil {
		panic(err)
	}

	return union
}

func MustParseInterface(i interface{}) *parser.Interface {
	interfaceType, err := ParseInterface(i)
	if err != nil {
		panic(err)
	}

	return interfaceType
}

func MustParseScalar(i interface{}) *parser.Scalar {
	scalar, err := ParseScalar(i)
	if err != nil {
		panic(err)
	}

	return scalar
}

func MustParseEnum(i interface{}) *parser.Enum {
	enum, err := ParseEnum(i)
	if err != nil {
		panic(err)
	}

	return enum
}
