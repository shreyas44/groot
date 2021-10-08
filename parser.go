package groot

import (
	"reflect"

	"github.com/shreyas44/groot/parser"
)

// wrapper around parser.ParseType
func ParseType(i interface{}) (*parser.Type, error) {
	return parser.NewType(reflect.TypeOf(i))
}

// same as ParseType, but panics if an error is encounters
func MustParseType(i interface{}) *parser.Type {
	t, err := ParseType(i)
	if err != nil {
		panic(err)
	}

	return t
}
