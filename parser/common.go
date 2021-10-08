package parser

import "reflect"

type ParserType interface {
	Kind() Kind
	ReflectType() reflect.Type
}

type ParserTypeWithFields interface {
	Fields() []*ObjectField
}
