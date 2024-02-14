package domain

import (
	"bytes"
	"encoding/base64"
	"log"
	"math/big"
	neturl "net/url"
	"strings"
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

// EncodedCursor is a scrambled string representation of the fields defined in DecodedCursor
type EncodedCursor string

// DecodedCursor the cursor values after decoding EncodedCursor
type DecodedCursor struct {
	FiltersChecksum []byte
	FID             int64
}

// PrevNextFID previous and next feature id (fid) to encode in cursor.
type PrevNextFID struct {
	Prev int64
	Next int64
}

// NewCursors create Cursors based on the prev/next feature ids from the datasource
// and the provided filters (captured in a hash).
func NewCursors(fid PrevNextFID, filtersChecksum []byte) Cursors {
	return Cursors{
		Prev: encodeCursor(fid.Prev, filtersChecksum),
		Next: encodeCursor(fid.Next, filtersChecksum),

		HasPrev: fid.Prev > 0,
		HasNext: fid.Next > 0,
	}
}

func encodeCursor(fid int64, filtersChecksum []byte) EncodedCursor {
	fidAsBytes := big.NewInt(fid).Bytes()

	// format of the cursor: <encoded fid><separator><encoded checksum>
	cursorB64 := base64.StdEncoding.EncodeToString(fidAsBytes) + string(separator) + base64.StdEncoding.EncodeToString(filtersChecksum)
	return EncodedCursor(neturl.QueryEscape(cursorB64))
}

// Decode turns encoded cursor into DecodedCursor and verifies the
// that the checksum of query params that act as filters hasn't changed
func (c EncodedCursor) Decode(filtersChecksum []byte) DecodedCursor {
	value, err := neturl.QueryUnescape(string(c))
	if err != nil || value == "" {
		return DecodedCursor{filtersChecksum, 0}
	}

	// split first, then decode
	encoded := strings.Split(value, string(separator))
	if len(encoded) < 2 {
		log.Printf("cursor '%s' doesn't contain expected separator %c", value, separator)
		return DecodedCursor{filtersChecksum, 0}
	}
	decodedFid, fidErr := base64.StdEncoding.DecodeString(encoded[0])
	decodedChecksum, checksumErr := base64.StdEncoding.DecodeString(encoded[1])
	if fidErr != nil || checksumErr != nil {
		log.Printf("decoding cursor value '%s' failed, defaulting to first page", value)
		return DecodedCursor{filtersChecksum, 0}
	}

	// feature id
	fid := big.NewInt(0).SetBytes(decodedFid).Int64()
	if fid < 0 {
		log.Printf("negative feature ID detected: %d, defaulting to first page", fid)
		fid = 0
	}

	// checksum
	if !bytes.Equal(decodedChecksum, filtersChecksum) {
		log.Printf("filters in query params have changed during pagination, resetting to first page")
		return DecodedCursor{filtersChecksum, 0}
	}

	return DecodedCursor{filtersChecksum, fid}
}

func (c EncodedCursor) String() string {
	return string(c)
}
