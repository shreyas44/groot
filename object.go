package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type Object struct {
	Name        string
	Description string
	object      *graphql.Object
	fields      []*Field
	interfaces  []*Interface
}

func (object *Object) GraphQLType() *graphql.Object {
	graphqlInterfaces := InterfacesFromInterfaces(object.interfaces)

	if object.object != nil {
		for _, field := range object.fields {
			object.object.AddFieldConfig(field.Name, field.GraphQLType())
		}

		return object.object
	}

	fields := graphql.Fields{}
	for _, field := range object.fields {
		fields[field.Name] = field.GraphQLType()
	}

	object.object = graphql.NewObject(graphql.ObjectConfig{
		Name:        object.Name,
		Fields:      fields,
		Description: object.Description,
		Interfaces:  graphqlInterfaces,
	})

	return object.object
}

func NewObject(t reflect.Type) *Object {
	// create resolvers for fields
	// For custom resolvers validate the argument types of the resolver

	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("must pass a reflect type of kind reflect.Struct, received %s", t.Kind()))
	}

	structName := t.Name()
	object := &Object{
		Name:       structName,
		fields:     []*Field{},
		interfaces: []*Interface{},
	}

	graphqlTypes[t] = object.GraphQLType()

	structFieldCount := t.NumField()
	for i := 0; i < structFieldCount; i++ {
		structField := t.Field(i)
		field := NewField(structField)

		// field is a relationship if it's nil
		if field != nil {
			object.fields = append(object.fields, field)
		}
	}

	object.GraphQLType()

	return object
}

func NewObjects(types ...reflect.Type) []*Object {
	objects := []*Object{}

	for _, t := range types {
		objects = append(objects, NewObject(t))
	}

	return objects
}
