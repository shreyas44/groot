# Unions

[Unions](https://graphql.org/learn/schema/#union-types) allow you to compose multiple types together. Let's say we have a search query that can either return a post or user. We can define a Union type like below to allow this.

```graphql
union SearchResult = Post | User
```

To define the Union in Groot, we need to create a new struct that embeds `groot.BaseType` and all the members of the Union.

```go
type SearchResult struct {
	groot.UnionType
	Post
	User
}
```

You can then use the type anywhere you like.

```go
type Query struct {
	Search SearchResult `json:"search"`
}

type SearchArgs struct {
	Query string `json:"query"`
}

type (q Query) ResolveSearch(args SearchArgs) (SearchResult, error) {
	searchResults, err := db.Search(args.Query)

	results := []SearchResult{}
	for _, result := range searchResults {
		switch result := result.(type) {
		case Post:
			results = append(results, SearchResult{Post: result})
		case User:
			results = append(results, SearchResult{User: result})
		}
	}

	return results, nil
}
```

If you're coming from another GraphQL implementation, you would notice there's no type resolver here to determine which type the union is returning. In Groot, **you don't need to define it manually** since the type resolution is done when you initialize the value, i.e. the type which we have to return will have a non zero value, and all others will have a zero value. For example, if we have `SearchResult{Post: result}`, `SearchResult.Post` has a non zero value while all other field have a zero value, which means we can resolve the type to `Post`.
