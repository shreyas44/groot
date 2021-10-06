package relay

import "errors"

type NodeDefinition struct {
	ID string `json:"id"`
}

type Node interface {
	ImplementsNode() NodeDefinition
}

func (d NodeDefinition) ImplementsNode() NodeDefinition {
	return d
}

type PageInfo struct {
	HasPreviousPage bool   `json:"hasPreviousPage"`
	HasNextPage     bool   `json:"hasNextPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

type PaginationArgs struct {
	First  int    `json:"first"`
	Last   int    `json:"last"`
	After  string `json:"after"`
	Before string `json:"before"`
}

func (args PaginationArgs) Validate() error {
	if args.First != 0 && args.Last != 0 {
		return errors.New("first and last cannot be set at the same time")
	}

	if args.First < 0 || args.Last < 0 {
		return errors.New("first and last must be greater than 0")
	}

	if args.After != "" && args.Before != "" {
		return errors.New("after and before cannot be set at the same time")
	}

	if args.First != 0 && args.Before != "" {
		return errors.New("first and before cannot be set at the same time")
	}

	if args.Last != 0 && args.After != "" {
		return errors.New("last and after cannot be set at the same time")
	}

	return nil
}

// We can add these once we have generics in Go

// type ConnectionEdge struct {
// 	Cursor string `json:"cursor"`
// 	Node   Node   `json:"node"`
// }

// type Connection struct {
// 	PageInfo PageInfo         `json:"pageInfo"`
// 	Edges    []ConnectionEdge `json:"edges"`
// }
