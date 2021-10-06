# Nullables

To create nullable types, use a pointer instead of a value type. This desing choice was made since it makes composing nullable types a lot easier.

```go
`[[String!]!]!` ->     `[][]string`
`[[String]!]!`  ->    `[][]*string`
`[[String]]!`   ->   `[]*[]*string`
`[[String]]`    ->  `*[]*[]*string`
```

The same applied for non-scalar types as well.
