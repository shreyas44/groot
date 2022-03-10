package parser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ExampleArrayElem struct{}

func TestParsedArray(t *testing.T) {
	var (
		stringListType = reflect.TypeOf([]string{})
		stringType     = reflect.TypeOf("")
		structListType = reflect.TypeOf([]ExampleArrayElem{})
		structElem     = structListType.Elem()

		testCases = []struct {
			name         string
			isArg        bool
			reflectTyp   reflect.Type
			expectedType Type
		}{
			{"FieldWithStructElement", false, structListType, &Array{structListType, &Object{structElem, []*Field{}, []*Interface{}}}},
			{"ArgWithStructElement", true, structListType, &Array{structListType, &Input{structElem, nil, []*Argument{}}}},
			{"FieldWithStringElement", false, stringListType, &Array{stringListType, &Scalar{stringType}}},
			{"ArgWithStringElement", true, stringListType, &Array{stringListType, &Scalar{stringType}}},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			typ, err := NewArray(testCase.reflectTyp, testCase.isArg)
			require.Nil(t, err)
			assert.Equal(t, testCase.expectedType, typ, testCase)
		})
	}
}
