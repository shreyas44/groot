package parser

import (
	"fmt"
	"reflect"
)

type Union struct {
	reflect.Type
	members []*Object
}

func NewUnion(t reflect.Type) (*Union, error) {
	union := &Union{
		Type:    t,
		members: []*Object{},
	}

	if err := validateTypeKind(union); err != nil {
		panic(err)
	}

	cache.set(t, union)

	if err := validateUnion(union); err != nil {
		return nil, err
	}

	for i := 0; i < t.NumField(); i++ {
		embeddedStruct := t.Field(i).Type

		if embeddedStruct == reflect.TypeOf(UnionType{}) {
			continue
		}

		member, err := getOrCreateType(embeddedStruct)
		if err != nil {
			return nil, err
		}

		union.members = append(union.members, member.(*Object))
	}

	return union, nil
}

func (u *Union) Members() []*Object {
	return u.members
}

func (u *Union) ReflectType() reflect.Type {
	return u.Type
}

func validateUnion(t *Union) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		parserType, err := getTypeKind(field.Type)
		if err != nil {
			return err
		}

		if parserType != KindObject && !field.Anonymous {
			return fmt.Errorf(
				"got extra field %s on union %s, union types cannot contain any field other than embedded structs and groot.UnionType",
				field.Name,
				t.Name(),
			)
		}
	}

	return nil
}
