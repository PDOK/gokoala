package domain

import (
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
