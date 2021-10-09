package parser

import "reflect"

type Object struct {
	reflect.Type
	fields     []*Field
	interfaces []*Interface
}

func NewObject(t reflect.Type) (*Object, error) {
	object := &Object{
		Type:       t,
		fields:     []*Field{},
		interfaces: []*Interface{},
	}

	if err := validateTypeKind(t, KindObject); err != nil {
		panic(err)
	}

	cache.set(t, object)

	fields, err := getFields(object, t)
	if err != nil {
		return nil, err
	}

	interfaces, err := getInterfaces(object)
	if err != nil {
		return nil, err
	}

	object.fields = fields
	object.interfaces = interfaces

	return object, nil
}

func (o *Object) Fields() []*Field {
	return o.fields
}

func (o *Object) Interfaces() []*Interface {
	return o.interfaces
}

func (o *Object) ReflectType() reflect.Type {
	return o.Type
}

func getFields(t TypeWithFields, reflectType reflect.Type) ([]*Field, error) {
	fields := []*Field{}

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)

		if field.Anonymous {
			embeddedFields, err := getFields(t, field.Type)
			if err != nil {
				return nil, err
			}

			fields = append(fields, embeddedFields...)
			continue
		}

		objectField, err := NewField(t, field)
		if err != nil {
			return nil, err
		}

		if objectField != nil {
			fields = append(fields, objectField)
		}
	}

	return fields, nil
}

func getInterfaces(object *Object) ([]*Interface, error) {
	interfaces := []*Interface{}

	for i := 0; i < object.Type.NumField(); i++ {
		field := object.Type.Field(i)

		if field.Anonymous {
			interfaceType, err := getOrCreateType(field.Type)
			if err != nil {
				return nil, err
			}

			if interfaceType, ok := interfaceType.(*Interface); ok {
				interfaces = append(interfaces, interfaceType)
			}
		}
	}

	return interfaces, nil
}
