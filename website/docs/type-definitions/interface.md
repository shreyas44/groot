# Interfaces

[Interface](https://graphql.org/learn/schema/#interfaces) definitions have two parts, an `interface` and a `struct`. Let's see how we can implement the above GraphQL schema using Groot.

```graphql
interface Character {
  id: ID!
  name: String!
  friends: [Character]
  appearsIn: [Episode]!
}

type Droid implements Character {
  id: ID!
  name: String!
  friends: [Character]
  appearsIn: [Episode]!
  primaryFunction: String
}
```

First, we need to create a GraphQL interface definition _struct_ which embeds `groot.InterfaceType` and with name `CharacterDefinition`.

```go
type CharacterDefinition struct {
	groot.InterfaceType
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	Friends   []string `json:"fiends"`
	AppearsIn []string `json:"appearsIn"`
}
```

Next, we need to create an interface with name `Character` with a single method `ImplementsCharacter()` with a return type of `CharacterDefinition`.

```go
type Character interface {
	ImplementsCharacter() CharacterDefinition
}
```

Then, we create the `ImplementsCharacter` method for `CharacterDefinition`.

```go
func (c CharacterDefinition) ImplementsCharacter() CharacterDefinition {
	// this can even return an empty type, the value returned doesn't matter
	return c
}
```

Finally, we can implement the interface in `Droid` by embedding the `CharacterDefinition` struct like below:

```go
type Droid struct {
	CharacterDefinition
	PrimaryFunction string `json:"primaryFunction"`
}
```

In the end it would look like this:

```go
type Character interface {
	ImplementsCharacter() CharacterDefinition
}

type CharacterDefinition struct {
	groot.InterfaceType
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	Friends   []string `json:"fiends"`
	AppearsIn []string `json:"appearsIn"`
}

func (c CharacterDefinition) ImplementsCharacter() CharacterDefinition {
	return c
}

type Droid struct {
	CharacterDefinition
	PrimaryFunction string `json:"primaryFunction"`
}
```

_Note, if we didn't create the `ImplementsCharacter` method on `CharacterDefinition`, we would have to implement it for each type that implements `Character`. Implementing it once for `CharacterDefinition` makes everyone's life easier._

To use the interface as the type on a field, we would use the `Character` type.

```go
type Query struct {
	Character Character `json:"character"`
}
```

Providing `CharacterDefinition` instead of `Character` anywhere will cause Groot to panic.

### Passing Type Directly to Schema

Sometimes you may have to pass a type that implements an interface directly to the schema config. This is required when you haven't used the type directly anywhere. For example, if we didn't use `Droid` anywhere above, we would have to pass it to the schema config like below:

```go
schema := groot.NewSchema(groot.SchemaConfig{
	Query: reflect.TypeOf(Query{}),
	Types: []reflect.Type(
		reflect.TypeOf(Droid{}),
	)
})
```

It might be a good practice to pass all types that implement an interface to the schema config to avoid any issues.
