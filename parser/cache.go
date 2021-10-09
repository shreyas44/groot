package parser

import "reflect"

type typeCache map[reflect.Type]Type

var cache = typeCache{}

func (c typeCache) get(t reflect.Type) (Type, bool) {
	parserType, ok := c[t]
	return parserType, ok
}

func (c typeCache) set(t reflect.Type, parserType Type) {
	c[t] = parserType
}
