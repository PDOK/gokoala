package domain

import (
	"log"

	"github.com/sqids/sqids-go"
)

const (
	cursorAlphabet = "1Vti5BYcjOdTXunDozKPm4syvG6galxLM8eIrUS2bWqZCNkwpR309JFAHfh7EQ" // generated on https://sqids.org/playground
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

	HasPrev bool
	HasNext bool
}

type NextPrevID struct {
	Prev int64
	Next int64
}

func NewCursor(features []*Feature, id NextPrevID) Cursor {
	limit := len(features)
	if limit == 0 {
		return Cursor{}
	}

	return Cursor{
		Prev: encodeCursor(uint64(id.Prev)),
		Next: encodeCursor(uint64(id.Next)),

		HasPrev: id.Prev > 0,
		HasNext: id.Next > 0,
	}
}

// EncodedCursor is a scrambled string representation of a consecutive ordered integer cursor
type EncodedCursor string

func encodeCursor(value uint64) EncodedCursor {
	encodedValue, err := cursorCodec.Encode([]uint64{value})
	if err != nil {
		log.Printf("failed to encode cursor value %d", value)
		return ""
	}
	return EncodedCursor(encodedValue)
}

// Decode turn encoded cursor string into cursor value and orderBy direction
func (c EncodedCursor) Decode() int64 {
	value := string(c)
	if value == "" {
		return 0
	}
	decodedValue := cursorCodec.Decode(value)
	if len(decodedValue) != 1 {
		log.Printf("expected 1 value after decoding, but received: '%v'", decodedValue)
	} else if len(decodedValue) == 0 {
		log.Printf("decoding cursor value '%v' failed, defaulting to first page", decodedValue)
		return 0
	}

	cursor := int64(decodedValue[0])
	if cursor < 0 {
		cursor = 0
	}
	return cursor
}
