package domain

import (
	"math"
	"reflect"
	"testing"
)

func TestNewCursor(t *testing.T) {
	type args struct {
		features []*Feature
		id       PrevNextFID
	}
	var tests = []struct {
		name string
		args args
		want Cursors
	}{
		{
			name: "test first page",
			args: args{
				features: []*Feature{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}},
				id: PrevNextFID{
					Prev: 0,
					Next: 4,
				},
			},
			want: Cursors{
				Prev:    "fA==",
				Next:    "BHw=",
				HasPrev: false,
				HasNext: true,
			},
		},
		{
			name: "test last page",
			args: args{
				features: []*Feature{{ID: 5}, {ID: 6}, {ID: 7}, {ID: 8}},
				id: PrevNextFID{
					Prev: 4,
					Next: 0,
				},
			},
			want: Cursors{
				Prev:    "BHw=",
				Next:    "fA==",
				HasPrev: true,
				HasNext: false,
			},
		},
		{
			name: "test middle page",
			args: args{
				features: []*Feature{{ID: 3}, {ID: 4}, {ID: 5}, {ID: 6}},
				id: PrevNextFID{
					Prev: 2,
					Next: 7,
				},
			},
			want: Cursors{
				Prev:    "Anw=",
				Next:    "B3w=",
				HasPrev: true,
				HasNext: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCursors(tt.args.id, []byte{})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCursors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodedCursor_Decode(t *testing.T) {
	type args struct {
		filtersChecksum []byte
	}
	tests := []struct {
		name string
		c    EncodedCursor
		args args
		want DecodedCursor
	}{
		{
			name: "should return cursor if no checksum is available in cursor, and no expected checksum provided",
			c:    encodeCursor(123, []byte{}),
			args: args{
				filtersChecksum: []byte{},
			},
			want: DecodedCursor{
				FID:             123,
				FiltersChecksum: []byte{},
			},
		},
		{
			name: "should not fail on checksum which contains separator",
			c:    encodeCursor(123456, []byte{'a', separator, 'b'}),
			args: args{
				filtersChecksum: []byte{'a', separator, 'b'},
			},
			want: DecodedCursor{
				FID:             123456,
				FiltersChecksum: []byte{'a', separator, 'b'},
			},
		},
		{
			name: "should not fail on checksum which contains only separator",
			c:    encodeCursor(123456, []byte{separator}),
			args: args{
				filtersChecksum: []byte{separator},
			},
			want: DecodedCursor{
				FID:             123456,
				FiltersChecksum: []byte{separator},
			},
		},
		{
			name: "should fail (return 0 fid) on non matching checksums",
			c:    encodeCursor(123456, []byte("foobarbaz")),
			args: args{
				filtersChecksum: []byte("bazbar"),
			},
			want: DecodedCursor{
				FID:             0,
				FiltersChecksum: []byte("bazbar"),
			},
		},
		{
			name: "should handle large feature id",
			c:    encodeCursor(math.MaxInt64, []byte("foobar")),
			args: args{
				filtersChecksum: []byte("foobar"),
			},
			want: DecodedCursor{
				FID:             math.MaxInt64,
				FiltersChecksum: []byte("foobar"),
			},
		},
		{
			name: "should always return positive feature id",
			c:    encodeCursor(math.MinInt64, []byte("foobar")),
			args: args{
				filtersChecksum: []byte("foobar"),
			},
			want: DecodedCursor{
				FID:             0,
				FiltersChecksum: []byte("foobar"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Decode(tt.args.filtersChecksum); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}
