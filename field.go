package groot

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
)

type GrootField struct {
	t GrootType
	field reflect.StructField
}

// Get an instance of graphql.Field
func (f *GrootField) Field() graphql.Field {
	return graphql.Field{
		Type: f.Type(),
		Resolve: f.Resolver(),
		DeprecationReason: f.field.Tag.Get("deprecate"),
		Description: f.field.Tag.Get("description"),
	}
}

// Get the graphql.Type for field
func (f *GrootField) Type() graphql.Type {
	return types[f.field.Type.Name()]
}

// Get the resolver for field. Gets default resovler if no custom resolver is set
func (f *GrootField) Resolver() graphql.FieldResolveFn {
	s := f.t._struct
	resolver, ok := s.MethodByName("Resolve" + f.field.Type.Name())

	if !ok {
		// default resolver
		return func(p graphql.ResolveParams) (interface{}, error) {
			return reflect.ValueOf(p.Source).FieldByName(f.field.Type.Name()).Interface(), nil
		}
	}

	// check resolver type

	numOut := resolver.Func.Type().NumOut()
	if numOut != 2 {
		msg := fmt.Sprintf(
			"resolver for %v must have the return 2 values of type (%v, error), returned only %d value", 
			f.field.Type.Name(), 
			f.field.Type.Name(), 
			numOut,
		)
		panic(msg)
	}

	fReturnType := resolver.Func.Type().Out(0)
	sReturnType := resolver.Func.Type().Out(1)

	if fReturnType != f.field.Type || sReturnType != reflect.TypeOf(new(error)) {
		msg := fmt.Sprintf(
			"resolver for %v must have a return type of (%v, error), got (%v, %v)", 
			f.field.Type.Name(), 
			f.field.Type.Name(), 
			fReturnType, 
			sReturnType,
		)

		panic(msg)
	}

	return func(p graphql.ResolveParams) (interface{}, error) {
		args := []reflect.Value{}

		// add context to args
		args = append(args, reflect.ValueOf(p.Context))

		// add input args to args
		response := resolver.Func.Call(args)

		return response[0], response[1].Interface().(error)
	}
}

// Get the field name of field
func (f *GrootField) Name(config ObjectConfig) string {
	if name := f.field.Tag.Get("json"); name != "" {
		return name
	} 

	name := f.field.Type.Name()

	if config.CamelCase {
		name = strings.ToLower(string(name[0])) + name[1:]
	}
	
	return name
}

func NewField(t reflect.Type) *GrootField {
	return nil
}