package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type Scalar struct {
	scalar      *graphql.Scalar
	reflectType reflect.Type
}

func (scalar *Scalar) GraphQLType() graphql.Type {
	return scalar.scalar
}

func NewScalar(t reflect.Type, builder *SchemaBuilder) *Scalar {
	scalars := map[reflect.Kind]*graphql.Scalar{
		reflect.Int:     graphql.Int,
		reflect.String:  graphql.String,
		reflect.Bool:    graphql.Boolean,
		reflect.Float32: graphql.Float,
	}

	scalar := &Scalar{
		scalar:      scalars[t.Kind()],
		reflectType: t,
	}

	builder.types[t.Name()] = scalar.scalar
	builder.grootTypes[t] = scalar
	return scalar
}
