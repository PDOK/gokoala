package domain

import (
	"log"

	"github.com/sqids/sqids-go"
)

const (
	cursorAlphabet = "1Vti5BYcjOdTXunDozKPm4syvG6galxLM8eIrUS2bWqZCNkwpR309JFAHfh7EQ" // generated on https://sqids.org/playground

	OrderAsc     = "asc"
	OrderDesc    = "desc"
	orderAscInt  = 1
	orderDescInt = 2
)

var (
	cursorCodec, _ = sqids.New(sqids.Options{
		Alphabet:  cursorAlphabet,
		Blocklist: nil, // disable blocklist
		MinLength: 8,
	})
)

// Cursor since we use cursor-based pagination as opposed to offset-based pagination
type Cursor struct {
	Prev EncodedCursor
	Next EncodedCursor

	IsFirst bool
	IsLast  bool
}

func NewCursor(features []*Feature, last bool) Cursor {
	limit := len(features)
	if limit == 0 {
		return Cursor{}
	}

	prev := features[0].ID
	next := features[limit-1].ID

	return Cursor{
		Prev: encodeCursor(prev, true),
		Next: encodeCursor(next, false),

		IsFirst: next <= int64(limit),
		IsLast:  last,
	}
}

// EncodedCursor is a scrambled string representation of a consecutive ordered integer cursor
type EncodedCursor string

func encodeCursor(value int64, isPrev bool) EncodedCursor {
	// since sqids can only contain numbers, so we use integers instead of 'asc'/'desc' strings
	var orderBy uint64
	if isPrev {
		orderBy = orderDescInt
	} else {
		orderBy = orderAscInt
	}

	encodedValue, err := cursorCodec.Encode([]uint64{uint64(value), orderBy})
	if err != nil {
		log.Printf("failed to encode cursor value %d", value)
		return ""
	}
	return EncodedCursor(encodedValue)
}

// Decode turn encoded cursor string into cursor value and orderBy direction
func (c EncodedCursor) Decode() (int64, string) {
	value := string(c)
	if value == "" {
		return 0, OrderAsc
	}
	decodedValue := cursorCodec.Decode(value)
	if len(decodedValue) != 2 {
		log.Printf("expected 2 values after decoding, but received: '%v'", decodedValue)
	} else if len(decodedValue) == 0 {
		log.Printf("decoding cursor value '%v' failed, defaulting to first page", decodedValue)
		return 0, OrderAsc
	}

	cursor := int64(decodedValue[0])
	if cursor < 0 {
		cursor = 0
	}

	var orderBy string
	switch {
	case decodedValue[1] == orderAscInt:
		orderBy = OrderAsc
	case decodedValue[1] == orderDescInt:
		orderBy = OrderDesc
	default:
		log.Printf("invalid order by value received %d, defaulting to %s", decodedValue[1], OrderAsc)
		orderBy = OrderAsc
	}

	return cursor, orderBy
}
