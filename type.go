package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type GrootType struct {
	fields []GrootField
	_struct reflect.Type
}

// Get the graphql.Type instance for type
func (t *GrootType) Type() graphql.Type {
	// return types[t._struct.Name()]
	return nil
}

// Add field to type
func (t *GrootType) AddField(field GrootField) {
}

func NewType(t reflect.Type) *GrootType {
	return nil
}
