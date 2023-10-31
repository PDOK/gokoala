package domain

import (
	"bytes"
	"encoding/base64"
	"log"
	"math/big"
)

const separator = '|'

// Cursors holds next and previous cursor. Note that we use
// 'cursor-based pagination' as opposed to 'offset-based pagination'
type Cursors struct {
	Prev EncodedCursor
	Next EncodedCursor

	HasPrev bool
	HasNext bool
}

// EncodedCursor is a scrambled string representation of:
// - a consecutive ordered integer feature ID
// - a hash of the filters (limit, bbox, CQL filters, etc) used when querying features
type EncodedCursor string

// DecodedCursor the cursor values after decoding EncodedCursor
type DecodedCursor struct {
	ID          int64
	FiltersHash []byte
}

// PrevNextID id of previous and next feature id (fid) to encode in cursor.
type PrevNextID struct {
	Prev int64
	Next int64
}

// NewCursors create Cursors based on the prev/next feature ids from the datasource
// and the provided filters (captured in a hash).
func NewCursors(id PrevNextID, filtersHash []byte) Cursors {
	return Cursors{
		Prev: encodeCursor(id.Prev, filtersHash),
		Next: encodeCursor(id.Next, filtersHash),

		HasPrev: id.Prev > 0,
		HasNext: id.Next > 0,
	}
}

func encodeCursor(id int64, filtersHash []byte) EncodedCursor {
	// format of the cursor: <id>|<hash>
	cursorToEncode := append([]byte{byte(id), byte(separator)}, filtersHash...)

	encoded := base64.URLEncoding.EncodeToString(cursorToEncode)
	return EncodedCursor(encoded)
}

// Decode turn encoded cursor string into DecodedCursor and
// verify the 'filtersHash' hasn't changed
func (c EncodedCursor) Decode(filtersHash []byte) DecodedCursor {
	value := string(c)
	if value == "" {
		return DecodedCursor{0, filtersHash}
	}
	decoded, err := base64.URLEncoding.DecodeString(value)
	if err != nil || len(decoded) == 0 {
		log.Printf("decoding cursor value '%v' failed, defaulting to first page", decoded)
		return DecodedCursor{0, filtersHash}
	}
	parts := bytes.Split(decoded, []byte{separator})
	if len(decoded) < 1 {
		return DecodedCursor{0, filtersHash}
	}
	cursor := big.NewInt(0).SetBytes(parts[0]).Int64()
	if err != nil {
		log.Printf("cursor %s doesn't contain numeric value, defaulting to first page", parts[0])
		return DecodedCursor{0, filtersHash}
	}
	if cursor < 0 {
		cursor = 0
	}

	if len(parts) > 1 && bytes.Compare(parts[1], filtersHash) != 0 {
		log.Printf("filters (query params) changed during pagination, resetting to first page")
		return DecodedCursor{0, filtersHash}
	}

	return DecodedCursor{cursor, filtersHash}
}
