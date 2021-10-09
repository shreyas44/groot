package parser

import (
	"encoding/json"
)

type UnionType struct{}

type InterfaceType struct{}

type EnumType interface {
	Values() []string
}

type ScalarType interface {
	json.Marshaler
	json.Unmarshaler
}
