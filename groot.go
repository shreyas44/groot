package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

var (
	graphqlTypes map[reflect.Type]graphql.Type
	types        map[string]*graphql.Type
)

func init() {
	var (
		gstring    = ""
		gint       = 0
		gfloat     = float32(0.0)
		gbool      = true
		nullString = &gstring
		nullInt    = &gint
		nullFloat  = &gfloat
		nullBool   = &gbool
	)

	graphqlTypes = map[reflect.Type]graphql.Type{
		reflect.TypeOf(gstring):    graphql.NewNonNull(graphql.String),
		reflect.TypeOf(gint):       graphql.NewNonNull(graphql.Int),
		reflect.TypeOf(gfloat):     graphql.NewNonNull(graphql.Float),
		reflect.TypeOf(gbool):      graphql.NewNonNull(graphql.Boolean),
		reflect.TypeOf(nullString): graphql.String,
		reflect.TypeOf(nullInt):    graphql.Int,
		reflect.TypeOf(nullFloat):  graphql.Float,
		reflect.TypeOf(nullBool):   graphql.Boolean,
	}
}
