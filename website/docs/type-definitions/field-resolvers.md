# Field Resolvers

We'll be working with the below `Post` struct for this section.

```go
type Post struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Author   User      `json:"author"`
	Comments []Comment `json:"comments"`
	// this field will not be in the graphql schema
	AuthorID string    `json:"-"`
}
```

### Resolver Method

The resolver for a field is defined the method on the struct with name `Resolve{field-name}`. The resolvers return type must be `({FieldType}, error)`, and the method should be defined on the value struct (`Post`) and not the pointer of the sturct (`*Post`).

The method signature of the resolver can be any of the following:

1. `()` - No Arguments
2. `(args ArgsStruct)` - Arguments to accept from the API
3. `(ctx context.Context)` - Context of a request
4. `(info graphql.ResolverInfo)` - Info about the GraphQL request
5. `(args ArgsStruct, ctx context.Context)`
6. `(args ArgsStruct, info graphql.ResolverInfo)`
7. `(ctx context.Context, info graphql.ResolverInfo)`
8. `(args ArgsStruct, ctx context.Context, info graphql.ResolverInfo)`

For example, we can define a resolver for `Author` like below:

```go
func (post Post) ResolveAuthor(ctx context.Context) (User, error) {
	loader := loader.UserLoaderFromCtx(ctx)
	user, err := loader.Load(post.AuthorID)
	return user, err
}
```

**Don't worry about making a mistake with the argument or return types, Groot will almost always catch it and panic.**

### Accepting Arguments

Say we want to accept the argument `first`, and `after` for the `comments` field for pagintion. First, we need to define a struct with the arguments.

```go
type PaginationArgs struct {
	First string `json:"first"`
	After string `json:"after"`
}
```

Next, we can accept these arguments in the resolver method like below:

```go
func (post Post) ResolveComments(args PaginationArgs) ([]Comment, error) {
	comments, err := db.GetPostComments(post.Id, args.First, args.After)
	return comments, err
}
```

You can also accept nested structures like below:

```go
type BarInput struct {
	BarText string `json:"barText"`
}

type FooArgs struct {
	FooText string   `json:"fooText"`
	Bar     BarInput `json:"bar"`
}

func (m Mutation) ResolveNewFoo(args FooArgs) (Foo, error) {
	return db.CreateFoo(args)
}
```

For the above example, Groot will create an [input type](https://graphql.org/learn/schema/#input-types) named `BarInput` and reference that in the argument field type.

<!-- ### Context -->
