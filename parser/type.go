package parser

import (
	"fmt"
	"reflect"
)

type Kind int

const (
	Scalar Kind = iota
	CustomScalar
	Object
	Interface
	InterfaceDefinition
	Union
	Enum
	List
	Nullable
	InvalidType
)

func (kind Kind) String() string {
	typeMap := map[Kind]string{
		Scalar:              "Scalar",
		CustomScalar:        "CustomScalar",
		Object:              "Object",
		Interface:           "Interface",
		InterfaceDefinition: "InterfaceDefinition",
		Union:               "Union",
		Enum:                "Enum",
		List:                "List",
		Nullable:            "Nullable",
		InvalidType:         "InvalidType",
	}

	return typeMap[kind]
}

type Type struct {
	reflect.Type
	kind Kind

	// only for objects and interfaces
	fields []*ObjectField

	// only for objects
	interfaces []*Type

	// only for interface
	definition *Type

	// only for unions
	members []*Type

	// only for lists and nullables
	element *Type
}

func getOrCreateType(reflectType reflect.Type) (*Type, error) {
	parserType, ok := cache.get(reflectType)
	if ok {
		return parserType, nil
	}

	parserType, err := NewType(reflectType)
	if err != nil {
		return nil, err
	}

	return parserType, nil
}

func NewType(reflectType reflect.Type) (*Type, error) {
	kind, err := getTypeKind(reflectType)
	if err != nil {
		return nil, err
	}

	parserType := &Type{
		Type:       reflectType,
		kind:       kind,
		fields:     []*ObjectField{},
		members:    []*Type{},
		interfaces: []*Type{},
	}

	cache.set(reflectType, parserType)

	switch kind {
	case Object:
		fields, err := getFields(parserType, reflectType)
		if err != nil {
			return nil, err
		}

		interfaces, err := getInterfaces(parserType)
		if err != nil {
			return nil, err
		}

		parserType.fields = fields
		parserType.interfaces = interfaces

	case InterfaceDefinition:
		fields, err := getFields(parserType, reflectType)
		if err != nil {
			return nil, err
		}

		parserType.fields = fields

	case Interface:
		if err := validateInterface(parserType); err != nil {
			return nil, err
		}

		interfaceDefReflectType := reflectType.Method(0).Type.Out(0)
		interfaceDef, err := getOrCreateType(interfaceDefReflectType)
		if err != nil {
			return nil, err
		}

		parserType.definition = interfaceDef

	case Union:
		if err := validateUnion(parserType); err != nil {
			return nil, err
		}

		for i := 0; i < reflectType.NumField(); i++ {
			embeddedStruct := reflectType.Field(i).Type

			if embeddedStruct == reflect.TypeOf(UnionType{}) {
				continue
			}

			member, err := getOrCreateType(embeddedStruct)
			if err != nil {
				return nil, err
			}

			parserType.members = append(parserType.members, member)
		}
	case Nullable, List:
		element, err := getOrCreateType(reflectType.Elem())
		if err != nil {
			return nil, err
		}

		parserType.element = element
	}

	return parserType, nil
}

func (t Type) Kind() Kind {
	return t.kind
}

// panics if Type.Kind() != parser.Object or parser.InterfaceDefinition
func (t Type) Fields() []*ObjectField {
	if t.kind != Object && t.kind != InterfaceDefinition {
		panic("parser: cannot get fields of non-object type")
	}

	return t.fields
}

// panics if Type.Kind() != parser.Interface
func (t Type) Definition() *Type {
	if t.kind != Interface {
		panic("parser: cannot get definition of non-interface type")
	}

	return t.definition
}

// panics if Type.Kind() != parser.Union
func (t Type) Members() []*Type {
	if t.kind != Union {
		panic("parser: cannot get members of non-union type")
	}

	return t.members
}

// panics if Type.Kind() != parser.Object
func (t Type) Interfaces() []*Type {
	if t.kind != Object {
		panic("parser: cannot check if non-object type implements an interface")
	}

	return t.interfaces
}

// panics if Type.Kind() != parser.List or parser.Nullable
func (t Type) Element() *Type {
	if t.kind != List && t.kind != Nullable {
		panic("parser: cannot get element of non-list or nullable type")
	}

	return t.element
}

func isTypeUnion(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous && field.Type == reflect.TypeOf(UnionType{}) {
			return true
		}
	}

	return false
}

func isInterfaceDefinition(t reflect.Type) bool {
	interfaceType := reflect.TypeOf(InterfaceType{})

	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		if field := t.Field(i); field.Anonymous && field.Type == interfaceType {
			return true
		}
	}

	return false
}

func getTypeKind(t reflect.Type) (Kind, error) {
	var (
		enumType   = reflect.TypeOf((*EnumType)(nil)).Elem()
		scalarType = reflect.TypeOf((*ScalarType)(nil)).Elem()
	)

	if ptrT := reflect.PtrTo(t); ptrT.Implements(scalarType) {
		return CustomScalar, nil
	}

	switch t.Kind() {
	case reflect.Ptr:
		return Nullable, nil

	case reflect.Slice:
		return List, nil

	case reflect.Interface:
		return Interface, nil

	case reflect.Struct:
		if isTypeUnion(t) {
			return Union, nil
		}

		if isInterfaceDefinition(t) {
			return InterfaceDefinition, nil
		}

		return Object, nil

	case
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Float32, reflect.Float64,
		reflect.Bool:
		return Scalar, nil

	case reflect.String:
		if t.Name() == "string" || !t.Implements(enumType) {
			return Scalar, nil
		}

		return Enum, nil
	}

	return InvalidType, fmt.Errorf("couldn't parse type %s", t.Name())
}

func validateInterface(t *Type) error {
	if t.NumMethod() != 1 {
		return fmt.Errorf(
			"interface %s can have only one method",
			t.Name(),
		)
	}

	method := t.Method(0)

	if method.Type.NumIn() != 0 {
		return fmt.Errorf(
			"method %s on interface %s should not have input arguments",
			method.Name,
			t.Name(),
		)
	}

	if method.Type.NumOut() != 1 {
		return fmt.Errorf(
			"method %s on interface %s should return exactly one value",
			method.Name,
			t.Name(),
		)
	}

	interfaceDefinition, err := getOrCreateType(method.Type.Out(0))
	if err != nil {
		return err
	}

	if interfaceDefinition.Kind() != InterfaceDefinition {
		return fmt.Errorf(
			"method %s on interface %s should return a struct with groot.InterfaceType embedded",
			method.Name,
			t.Name(),
		)
	}

	if t.Name()+"Definition" != interfaceDefinition.Name() {
		return fmt.Errorf(
			"method %s on interface %s should return a struct named %sDefinition",
			method.Name,
			t.Name(),
			t.Name(),
		)
	}

	return nil
}

func validateUnion(t *Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		parserType, err := getTypeKind(field.Type)
		if err != nil {
			return err
		}

		if parserType != Object && !field.Anonymous {
			return fmt.Errorf(
				"got extra field %s on union %s, union types cannot contain any field other than embedded structs and groot.UnionType",
				field.Name,
				t.Name(),
			)
		}
	}

	return nil
}

func getFields(object *Type, reflectType reflect.Type) ([]*ObjectField, error) {
	fields := []*ObjectField{}

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)

		if field.Anonymous {
			embeddedFields, err := getFields(object, field.Type)
			if err != nil {
				return nil, err
			}

			fields = append(fields, embeddedFields...)
			continue
		}

		objectField, err := NewObjectField(object, field)
		if err != nil {
			return nil, err
		}

		if objectField != nil {
			fields = append(fields, objectField)
		}
	}

	return fields, nil
}

func getInterfaces(object *Type) ([]*Type, error) {
	interfaces := []*Type{}

	for i := 0; i < object.Type.NumField(); i++ {
		field := object.Type.Field(i)

		if field.Anonymous {
			interfaceType, err := getOrCreateType(field.Type)
			if err != nil {
				return nil, err
			}

			if interfaceType.Kind() == InterfaceDefinition {
				interfaces = append(interfaces, interfaceType)
			}
		}
	}

	return interfaces, nil
}
