package encoding

import (
	"encoding/binary"
	"fmt"
	"math"

	"errors"

	geom2 "github.com/twpayne/go-geom"
	wkb2 "github.com/twpayne/go-geom/encoding/wkb"
	wkbcommmon2 "github.com/twpayne/go-geom/encoding/wkbcommon"
)

var (
	ErrInvalidEnvelopeType = errors.New("invalid envelope type")
	ErrInsufficientBytes   = errors.New("not enough bytes")
	ErrInvalidMagicNumber  = errors.New("invalid magic number")
)

type EnvelopeType uint8

// Magic is the magic number encode in the header. It should be 0x4750
var Magic = [2]byte{0x47, 0x50}

const (
	EnvelopeTypeNone    = EnvelopeType(0)
	EnvelopeTypeXY      = EnvelopeType(1)
	EnvelopeTypeXYZ     = EnvelopeType(2)
	EnvelopeTypeXYM     = EnvelopeType(3)
	EnvelopeTypeXYZM    = EnvelopeType(4)
	EnvelopeTypeInvalid = EnvelopeType(5)
)

// NumberOfElements that the particular Envelope Type will have.
func (et EnvelopeType) NumberOfElements() int {
	switch et { //nolint:exhaustive
	case EnvelopeTypeNone:
		return 0
	case EnvelopeTypeXY:
		return 4
	case EnvelopeTypeXYZ:
		return 6
	case EnvelopeTypeXYM:
		return 6
	case EnvelopeTypeXYZM:
		return 8
	default:
		return -1
	}
}

func (et EnvelopeType) String() string {
	str := "NONEXYZMXYMINVALID"
	switch et { //nolint:exhaustive
	case EnvelopeTypeNone:
		return str[0:4]
	case EnvelopeTypeXY:
		return str[4 : 4+2]
	case EnvelopeTypeXYZ:
		return str[4 : 4+3]
	case EnvelopeTypeXYM:
		return str[8 : 8+3]
	case EnvelopeTypeXYZM:
		return str[4 : 4+4]
	default:
		return str[11:]
	}
}

// HEADER FLAG LAYOUT
// 7 6 5 4 3 2 1 0
// R R X Y E E E B
// R Reserved for future use. (should be set to 0)
// X GeoPackageBinary type // Normal or extented
// Y empty geometry
// E Envelope type
// B ByteOrder
// http://www.geopackage.org/spec/#flags_layout
const (
	maskByteOrder        = 1 << 0
	maskEnvelopeType     = 1<<3 | 1<<2 | 1<<1
	maskEmptyGeometry    = 1 << 4
	maskGeoPackageBinary = 1 << 5
)

type headerFlags byte

func (hf headerFlags) String() string { return fmt.Sprintf("0x%02x", uint8(hf)) }

// Endian will return the encoded Endianess
func (hf headerFlags) Endian() binary.ByteOrder {
	if hf&maskByteOrder == 0 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

// Envelope returns the type of the envelope.
func (hf headerFlags) Envelope() EnvelopeType {
	et := uint8((hf & maskEnvelopeType) >> 1)
	if et >= uint8(EnvelopeTypeInvalid) {
		return EnvelopeTypeInvalid
	}
	return EnvelopeType(et)
}

// BinaryHeader is the gpkg header that accompainies every feature.
type BinaryHeader struct {
	// See: http://www.geopackage.org/spec/
	magic    [2]byte // should be 0x47 0x50  (GP in ASCII)
	version  uint8   // should be 0
	flags    headerFlags
	srsid    int32
	envelope []float64
}

// DecodeBinaryHeader decodes the data into the BinaryHeader
func DecodeBinaryHeader(data []byte) (*BinaryHeader, error) {
	if len(data) < 8 {
		return nil, ErrInsufficientBytes
	}

	var bh BinaryHeader
	bh.magic[0] = data[0]
	bh.magic[1] = data[1]
	bh.version = data[2]
	bh.flags = headerFlags(data[3])
	en := bh.flags.Endian()
	bh.srsid = int32(en.Uint32(data[4 : 4+4])) //nolint:gosec

	bytes := data[8:]
	et := bh.flags.Envelope()
	if et == EnvelopeTypeInvalid {
		return nil, ErrInvalidEnvelopeType
	}
	if et == EnvelopeTypeNone {
		return &bh, nil
	}
	num := et.NumberOfElements()
	// there are 8 bytes per float64 value and we need num of them.
	if len(bytes) < (num * 8) {
		return nil, ErrInsufficientBytes
	}

	bh.envelope = make([]float64, 0, num)
	for i := 0; i < num; i++ {
		bits := en.Uint64(bytes[i*8 : (i*8)+8])
		bh.envelope = append(bh.envelope, math.Float64frombits(bits))
	}
	if bh.magic[0] != Magic[0] || bh.magic[1] != Magic[1] {
		return &bh, ErrInvalidMagicNumber
	}
	return &bh, nil

}

// SRSID is the SRS id of the feature.
func (h *BinaryHeader) SRSID() int32 {
	if h == nil {
		return 0
	}
	return h.srsid
}

// Size is the size of the header in bytes.
func (h *BinaryHeader) Size() int {
	if h == nil {
		return 0
	}
	return (len(h.envelope) * 8) + 8
}

// StandardBinary is the binary encoding plus some metadata
// should be stored as a blob
type StandardBinary struct {
	Header   *BinaryHeader
	SRSID    int32
	Geometry geom2.T
}

func DecodeGeometry(bytes []byte) (*StandardBinary, error) {
	h, err := DecodeBinaryHeader(bytes)
	if err != nil {
		return nil, err
	}

	geo, err := wkb2.Unmarshal(bytes[h.Size():], wkbcommmon2.WKBOptionEmptyPointHandling(wkbcommmon2.EmptyPointHandlingNaN))
	if err != nil {
		return nil, err
	}
	return &StandardBinary{
		Header:   h,
		SRSID:    h.SRSID(),
		Geometry: geo,
	}, nil
}
