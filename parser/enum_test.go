package parser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ExampleEnumWithValues string
type ExampleEnumWithoutValues string

const (
	ExampleEnumWithValues_One ExampleEnumWithValues = "ONE"
	ExampleEnumWithValues_Two ExampleEnumWithValues = "TWO"
)

const (
	ExampleEnumWithoutValues_One ExampleEnumWithoutValues = "ONE"
	ExampleEnumWithoutValues_Two ExampleEnumWithoutValues = "TWO"
)

func (ExampleEnumWithValues) Values() []string {
	return []string{
		string(ExampleEnumWithValues_One),
		string(ExampleEnumWithValues_Two),
	}
}

func TestParsedEnumWithoutValues(t *testing.T) {
	defer func() { recover() }()

	reflectT := reflect.TypeOf(ExampleEnumWithoutValues(""))
	NewEnum(reflectT)

	t.Error("expected enum without Values method to panic")
}

func TestParsedEnumWithValues(t *testing.T) {
	reflectT := reflect.TypeOf(ExampleEnumWithValues(""))
	enum, err := NewEnum(reflectT)
	require.Nil(t, err)

	expectedEnum := &Enum{
		reflectT,
		[]string{
			string(ExampleEnumWithValues_One),
			string(ExampleEnumWithValues_Two),
		},
	}

	assert.Equal(t, expectedEnum, enum)
}
