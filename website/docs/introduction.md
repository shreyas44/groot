---
slug: /
---

# Introduction

Groot is an implementation of GraphQL in Go. It's built on top of [`github.com/graphql-go/graphql`](https://github.com/graphql-go/graphql), which means it should support most existing tooling built for it.

## Motivation

Go already has a couple of implementation of GraphQL, so why another one?

### Type Safety

Go is statically typed, and GraphQL is type safe, which means we don't need to and shouldn't use `interface{}` anywhere. A simple user struct with custom resolvers would look something like this. Although most type checking is done by Go, additional checks like [resolver](./type-definitions/field-resolvers) return types are done by Groot on startup to avoid type errors altogether.

```go
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

### Code First

Although the schema first aproach has its advantages, code first is arguably easier to maintain in the long run without having to deal with federated schemas and schema stitching. For more info check out [this blog post](https://blog.logrocket.com/code-first-vs-schema-first-development-graphql/).

When you work with Groot, in a way you're defining your schema first as well since you're defining the structure of your data (struct, interfaces, enums, etc) first.

### No Boilerplate and Code Duplication

The only thing we want to worry about is our types, resolvers, and business logic, nothing more. We also don't want to redeclare our types in Go as well as GraphQL, it can get cumbersome to maintain and keep track of.

### Simple To Use

Seriously, it is.

---

Groot achieves all this while still being compatible with extensions built for `github.com/graphql-go/graphql`, currently the most popular implementation of GraphQL in Go.

#### Features Not Supported but Coming Soon:

- [Custom Scalars](https://github.com/shreyas44/groot/issues/3)
- [Custom Directives](https://github.com/shreyas44/groot/issues/4)
- [Subscriptions](https://github.com/shreyas44/groot/issues/1)
- [Descriptions for type definitions](https://github.com/shreyas44/groot/issues/2)
- [Enum value description and deprecation](https://github.com/shreyas44/groot/issues/2)
