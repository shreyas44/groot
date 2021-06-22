package groot

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type queueItem struct {
	name string
	field reflect.StructField
	object *graphql.Object
}

type RelationQueue struct {
	items []queueItem
}

func (queue *RelationQueue) add(name string, field reflect.StructField, object *graphql.Object) {
	queue.items = append(queue.items, queueItem{
		name: name,
		field: field,
		object: object,
	})
}

func (queue *RelationQueue) dispatch() {
	for _, item := range queue.items {
		field := item.field
		gType := GetGType(field)

		gField := graphql.Field{
			Type: gType,
			Description: field.Tag.Get("description"),
			DeprecationReason: field.Tag.Get("deprecate"),
		}

		item.object.AddFieldConfig(item.name, &gField)
	}
}