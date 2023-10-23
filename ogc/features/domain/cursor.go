package domain

import (
	"log"
	"strconv"

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

	IsFirst bool
	IsLast  bool
}

func NewCursor(features []*Feature, last bool) Cursor {
	limit := len(features)
	if limit == 0 {
		return Cursor{}
	}

	start := features[0].ID
	end := features[limit-1].ID

	prev := start
	if prev != 0 {
		prev -= int64(limit + 1)
		if prev < 0 {
			prev = 0
		}
	}
	next := end

	return Cursor{
		Prev: encodeCursor(prev),
		Next: encodeCursor(next),

		IsFirst: next <= int64(limit),
		IsLast:  last,
	}
}

// EncodedCursor is a scrambled string representation of a consecutive ordered integer cursor
type EncodedCursor string

func encodeCursor(value int64) EncodedCursor {
	encodedValue, err := cursorCodec.Encode([]uint64{uint64(value)})
	if err != nil {
		log.Printf("failed to encode cursor value %d, defaulting to unencoded value.", value)
		return EncodedCursor(strconv.FormatInt(value, 10))
	}
	return EncodedCursor(encodedValue)
}

func (c EncodedCursor) Decode() int64 {
	value := string(c)
	if value == "" {
		return 0
	}
	decodedValue := cursorCodec.Decode(value)
	if len(decodedValue) > 1 {
		log.Printf("encountered more than one cursor value after decoding: '%v', "+
			"this is not allowed! Defaulting to first value.", decodedValue)
	} else if len(decodedValue) == 0 {
		log.Printf("decoding cursor value '%v' failed, defaulting to first page", decodedValue)
		return 0
	}
	result := int64(decodedValue[0])
	if result < 0 {
		result = 0
	}
	return result
}
