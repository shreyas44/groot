package groot

import (
	"fmt"
	"reflect"

	"github.com/shreyas44/groot/parser"
)

// wrapper around parser.ParseType
func ParseType(i interface{}) (parser.Type, error) {
	return parser.ParseType(reflect.TypeOf(i))
}

// same as ParseType, but panics if an error is encountered
func MustParseType(i interface{}) parser.Type {
	t, err := ParseType(i)
	if err != nil {
		panic(err)
	}

	return t
}

func ParseObject(i interface{}) (*parser.Object, error) {
	parserType, err := parser.ParseType(reflect.TypeOf(i))
	if err != nil {
		return nil, err
	}

	if object, ok := parserType.(*parser.Object); ok {
		return object, nil
	}

	return nil, fmt.Errorf("non object type passed to ParseObject")
}

func MustParseObject(i interface{}) *parser.Object {
	object, err := ParseObject(i)
	if err != nil {
		panic(err)
	}

	return object
}
