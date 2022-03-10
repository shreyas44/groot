package parser

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type InputWithValidValidator struct{}
type InputWithNoValidator struct{}
type InputWithInvalidValidator struct{}

func (InputWithValidValidator) Validate() error {
	return nil
}

// should return error
func (InputWithInvalidValidator) Validate() {}

type ArgValidatorTestInput struct {
	ArgWithValidator                          string
	ArgWithoutValidator                       string
	ArgWithInvalidValidatorNoReturn           string
	ArgWithInvalidValidatorGreaterReturnCount string
	ArgWithInvalidValidatorInvalidReturnType  string
	ArgWithInvalidValidatorNoArg              string
	ArgWithInvalidValidatorGreaterArgCount    string
	ArgWithInvalidValidatorInvalidArgType     string
}

func (ArgValidatorTestInput) ValidateArgWithValidator(stringArg string) error {
	return nil
}

func (ArgValidatorTestInput) ValidateArgWithInvalidValidatorNoReturn(stringArg string) {}

func (ArgValidatorTestInput) ValidateArgWithInvalidValidatorGreaterReturnCount(stringArg string) (string, error) {
	return "", nil
}

func (ArgValidatorTestInput) ValidateArgWithInvalidValidatorInvalidReturnType(stringArg string) string {
	return ""
}

func (ArgValidatorTestInput) ValidateArgWithInvalidValidatorNoArg() error {
	return nil
}

func (ArgValidatorTestInput) ValidateArgWithInvalidValidatorGreaterArgCount(stringArg string, intArg int) error {
	return nil
}

func (ArgValidatorTestInput) ValidateArgWithInvalidValidatorInvalidArgType(intArg int) error {
	return nil
}

func TestNewArgumentValidator(t *testing.T) {
	inputReflectTyp := reflect.TypeOf(ArgValidatorTestInput{})
	input := &Input{reflectType: inputReflectTyp}
	setup := func(fieldName string) (reflect.Method, *Argument, *ArgumentValidator, error) {
		field, _ := inputReflectTyp.FieldByName(fieldName)
		method, _ := inputReflectTyp.MethodByName("Validate" + fieldName)
		arg := &Argument{structField: field, input: input}

		validator, err := NewArgumentValidator(arg)
		return method, arg, validator, err
	}

	t.Run("ReturnsValidator", func(t *testing.T) {
		method, arg, validator, err := setup("ArgWithValidator")
		assert.Equal(t, method.Type, validator.reflectMethod.Type)
		assert.Equal(t, arg, validator.argument)
		assert.Nil(t, err)
	})

	t.Run("ReturnsNilValidatorNilErr", func(t *testing.T) {
		_, _, validator, err := setup("ArgWithoutValidator")
		assert.Nil(t, validator)
		assert.Nil(t, err)
	})

	t.Run("ReturnsErr", func(t *testing.T) {
		_, _, validator, err := setup("ArgWithInvalidValidatorNoReturn")
		assert.NotNil(t, err)
		assert.Nil(t, validator)
	})
}

func TestNewInputValidator(t *testing.T) {
	setup := func(typ interface{}) (reflect.Type, *Input, *InputValidator, error) {
		inputType := reflect.TypeOf(typ)
		input := &Input{reflectType: inputType}
		validator, err := NewInputValidator(input)
		return inputType, input, validator, err
	}

	t.Run("ReturnsValidator", func(t *testing.T) {
		inputType, input, validator, err := setup(InputWithValidValidator{})
		assert.Equal(t, inputType.Method(0).Type, validator.reflectMethod.Type)
		assert.Equal(t, input, validator.input)
		assert.Nil(t, err)
	})

	t.Run("ReturnsNilValidatorNilErr", func(t *testing.T) {
		_, _, validator, err := setup(InputWithNoValidator{})
		assert.Nil(t, validator)
		assert.Nil(t, err)
	})

	t.Run("ReturnsErr", func(t *testing.T) {
		_, _, validator, err := setup(InputWithInvalidValidator{})
		assert.Nil(t, validator)
		assert.NotNil(t, err)
	})
}

func TestValidateArgsValidator(t *testing.T) {
	inputTyp := reflect.TypeOf(ArgValidatorTestInput{})
	testCases := []struct {
		name        string
		fieldName   string
		expectedErr error
	}{
		{
			name:        "NoErr",
			fieldName:   "ArgWithValidator",
			expectedErr: nil,
		},
		{
			name:        "NoReturnValues",
			fieldName:   "ArgWithInvalidValidatorNoReturn",
			expectedErr: errors.New(""),
		},
		{
			name:        "TooManyReturnValues",
			fieldName:   "ArgWithInvalidValidatorGreaterReturnCount",
			expectedErr: errors.New(""),
		},
		{
			name:        "InvalidReturnType",
			fieldName:   "ArgWithInvalidValidatorInvalidReturnType",
			expectedErr: errors.New(""),
		},
		{
			name:        "NoArguments",
			fieldName:   "ArgWithInvalidValidatorNoArg",
			expectedErr: errors.New(""),
		},
		{
			name:        "TooManyArguments",
			fieldName:   "ArgWithInvalidValidatorGreaterArgCount",
			expectedErr: errors.New(""),
		},
		{
			name:        "InvalidArgumentType",
			fieldName:   "ArgWithInvalidValidatorInvalidArgType",
			expectedErr: errors.New(""),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			field, _ := inputTyp.FieldByName(tc.fieldName)
			method, _ := inputTyp.MethodByName("Validate" + tc.fieldName)

			input := &Input{reflectType: inputTyp}
			arg := &Argument{structField: field, input: input}

			_ = validateArgValidator(method, arg)
			// TODO: assert
			// assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestValidateInputValidator(t *testing.T) {

}
