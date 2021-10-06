# Objects

[Objects](https://graphql.org/learn/schema/#type-system) can be defined using a regular Go struct.

```go
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

For more info on field definitions, see [Field Definitions](./field-definitions).
