package parser

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ArgTestCustomScalar string
type ArgTestEnum string
type ArgTestEmptyInput struct{}
type ArgTestUnionMember1 struct{}
type ArgTestUnionMember2 struct{}

type ArgTestUnion struct {
	UnionType
	ArgTestUnionMember1
	ArgTestUnionMember2
}

type ArgTestInterfaceDefinition struct {
	InterfaceType
}

type ArgTestInterface interface {
	ImplementsArgTestInterface() ArgTestInterfaceDefinition
}

type ArgTestInput struct {
	StringArg      string `json:"stringArg"`
	NilJsonArg     string `json:"-"`
	ArgWithoutJSON string
	//lint:ignore U1000 argument is used through reflection
	unexportedArg string
}

const (
	ArgTestEnum_One ArgTestEnum = "One"
	ArgTestEnum_Two ArgTestEnum = "Two"
)

func (e ArgTestEnum) Values() []string {
	return []string{
		string(ArgTestEnum_One),
		string(ArgTestEnum_Two),
	}
}

func TestNewArgument(t *testing.T) {
	inputType := reflect.TypeOf(ArgTestInput{})
	stringType := reflect.TypeOf("")
	stringArg, _ := inputType.FieldByName("StringArg")
	input := &Input{reflectType: inputType}

	arg, err := NewArgument(input, stringArg)
	require.Nil(t, err)

	expectedArg := &Argument{
		input:       input,
		structField: stringArg,
		type_:       &Scalar{stringType},
		jsonName:    "stringArg",
	}

	assert.Equal(t, expectedArg, arg)

	t.Run("TestNilArgumentReturned", func(t *testing.T) {
		testCases := []struct {
			name      string
			fieldName string
		}{
			{
				name:      "WithUnexportedArg",
				fieldName: "unexportedArg",
			},
			{
				name:      "WithNilJsonArg",
				fieldName: "NilJsonArg",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				field, _ := inputType.FieldByName(testCase.fieldName)
				arg, err := NewArgument(input, field)
				assert.Nil(t, arg)
				assert.Nil(t, err)
			})
		}
	})

	t.Run("TestJSONName", func(t *testing.T) {
		testCases := []struct {
			name             string
			fieldName        string
			expectedJSONName string
		}{
			{
				name:             "WithJSONStructTag",
				fieldName:        "StringArg",
				expectedJSONName: "stringArg",
			},
			{
				name:             "WithoutJSONStructTag",
				fieldName:        "ArgWithoutJSON",
				expectedJSONName: "ArgWithoutJSON",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				field, _ := inputType.FieldByName(testCase.fieldName)
				arg, err := NewArgument(input, field)
				assert.Nil(t, err)
				assert.Equal(t, testCase.expectedJSONName, arg.JSONName())
			})
		}

	})
}

func TestGetOrCreateArgumentType(t *testing.T) {
	var (
		stringType         = reflect.TypeOf("")
		customScalarType   = reflect.TypeOf(ArgTestCustomScalar(""))
		enumType           = reflect.TypeOf(ArgTestEnum_One)
		listType           = reflect.TypeOf([]string{})
		structType         = reflect.TypeOf(ArgTestEmptyInput{})
		interfaceType      = reflect.TypeOf((*ArgTestInterface)(nil)).Elem()
		interfaceDefType   = reflect.TypeOf(ArgTestInterfaceDefinition{})
		nullableStringType = reflect.TypeOf((*string)(nil))
		unionType          = reflect.TypeOf(ArgTestUnion{})
		unsupportedErr     = errors.New("interface and union not supported for argument type")
		testCases          = []struct {
			name         string
			typ          reflect.Type
			expectedErr  error
			expectedType Type
		}{
			{
				name:         "Scalar",
				typ:          stringType,
				expectedErr:  nil,
				expectedType: &Scalar{stringType},
			},
			{
				name:         "CustomScalar",
				typ:          customScalarType,
				expectedErr:  nil,
				expectedType: &Scalar{customScalarType},
			},
			{
				name:         "Enum",
				typ:          enumType,
				expectedErr:  nil,
				expectedType: &Enum{enumType, ArgTestEnum_One.Values()},
			},
			{
				name:         "List",
				typ:          listType,
				expectedErr:  nil,
				expectedType: &Array{listType, &Scalar{stringType}},
			},
			{
				name:         "Input",
				typ:          structType,
				expectedErr:  nil,
				expectedType: &Input{structType, nil, []*Argument{}},
			},
			{
				name:         "Nullable",
				typ:          nullableStringType,
				expectedErr:  nil,
				expectedType: &Nullable{nullableStringType, &Scalar{stringType}},
			},
			{
				name:         "Interface",
				typ:          interfaceType,
				expectedErr:  unsupportedErr,
				expectedType: nil,
			},
			{
				name:         "InterfaceDefinition",
				typ:          interfaceDefType,
				expectedErr:  unsupportedErr,
				expectedType: nil,
			},
			{
				name:         "Union",
				typ:          unionType,
				expectedErr:  unsupportedErr,
				expectedType: nil,
			},
		}
	)

	t.Run("WithEmptyCache", func(t *testing.T) {
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				resetCache()
				typ, err := getOrCreateArgumentType(testCase.typ)
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedType, typ)
			})
		}
	})

	t.Run("WithCacheContainingFieldTypes", func(t *testing.T) {
		resetCache()

		// fill cache
		cache.set(stringType, &Scalar{stringType})
		cache.set(structType, &Object{structType, []*Field{}, []*Interface{}})
		cache.set(interfaceType, &Interface{interfaceType, []*Field{}})
		cache.set(unionType, &Union{unionType, []*Object{
			{reflect.TypeOf(ArgTestUnionMember1{}), []*Field{}, []*Interface{}},
			{reflect.TypeOf(ArgTestUnionMember2{}), []*Field{}, []*Interface{}},
		}})

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				typ, err := getOrCreateArgumentType(testCase.typ)
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedType, typ)
			})
		}
	})
}
