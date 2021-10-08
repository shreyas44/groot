# Subscriptions

Subscription type definitions are similar to the type definitions of regular objects. However, instead of having a `Resolver{FieldName}` method as a resolver, you have a `Subscribe{FieldName}` method that returns a receive only channel of the result type.

You can define arguments for a subscription field similar to how you would for a regular field resolver as discussed [here](./type-definitions/field-resolvers).

### Example with Timer

```go
package main

import (
	"time"

	"github.com/shreyas44/groot"
	"github.com/shreya44/handler"
)

type Query struct {
	Hello string `json:"hello"`
}

func (q Query) ResolveHello() (string, error) {
	return "Hello World", nil
}

type Notification struct {
	Time int `json:"time"`
}

type Subscription struct {
	Notification Notification `json:"notification"`
}

func (s Subscription) SubscribeNotification(ctx context.Context) (<-chan Notification, error) {
	ch := make(chan Notification)
	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				ch <- Notification{
					Time: int(time.Now().Unix()),
				}
			}
		}
	}()

	return ch, nil
}

func main() {
	schema, err := groot.NewSchema(groot.SchemaConfig{
		Query:        reflect.TypeOf(Query{}),
		Subscription: reflect.TypeOf(Subscription{}),
	})

	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	subscriptionHandler := handler.NewSubscriptionHandler(&schema)

	http.Handle("/graphql", h)
	http.Handle("/subscriptions", subscriptionHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

The [`github.com/shreyas44/handler`](https://github.com/shreyas44/handler) is a fork of the [`github.com/graphql-go/handler`](https://github.com/graphql-go/handler) package with an additional `NewSubscriptionHandler` function to handle subscriptions.
