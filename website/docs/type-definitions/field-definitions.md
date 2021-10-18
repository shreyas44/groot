# Field Definitions

The contents of this section apply to [Object](./object), [Input Object](./input), and [Interface](./interface) type definitions.

### Nullables and Arrays

Arrays can be defined using standard Go slices. For example, `[String!]!` would correspond to `[]string`.

By default, all types are non nullable. To create nullable types, use a pointer instead of a value type. This design choice was made since it makes composing nullable types a lot easier. This design can be revisited once generics are officially released in Go.

```go
`[[String!]!]!` ->     `[][]string`
`[[String]!]!`  ->    `[][]*string`
`[[String]]!`   ->   `[]*[]*string`
`[[String]]`    ->  `*[]*[]*string`
```

The same applies for non-scalar types as well.

### Field Descriptions

To add a description, define the struct tag `description` with the value being the description itself.

```go
type User struct {
	ID string `json:"id" description:"id of user"`
}
```

_Note, in the future we would use the docstring of the field as the description since it documents the field in Go as well as the API. However, you would still be able to override it with the description struct tag._

### Field Arguments

Refer [Resolvers](./field-resolvers#accepting-arguments) for more information on how to define arguments.

### Deprecating Fields

To deprecate a field, define the struct tag `deprecate` with the value being the reason you're deprecating the field.

```go
type User struct {
	OldID string `json:"oldId" deprecate:"old field"`
}
```

### Ignoring Fields

Ignoring fields allows us to pass values down the resolution tree without exposing them in the API. To ignore a field, you can either not export the field, or set the `json` struct tag to `-`.

```go
type User struct {
	otherPassword string
	Password      string `json:"-"`
}
```
