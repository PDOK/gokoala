package domain

import (
	"reflect"
	"testing"
)

func TestNewCursor(t *testing.T) {
	type args struct {
		features []*Feature
		id       PrevNextID
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
				id: PrevNextID{
					Prev: 0,
					Next: 4,
				},
			},
			want: Cursors{
				Prev:    "1GpOCgaM",
				Next:    "eVc7GU6Q",
				HasPrev: false,
				HasNext: true,
			},
		},
		{
			name: "test last page",
			args: args{
				features: []*Feature{{ID: 5}, {ID: 6}, {ID: 7}, {ID: 8}},
				id: PrevNextID{
					Prev: 4,
					Next: 0,
				},
			},
			want: Cursors{
				Prev:    "eVc7GU6Q",
				Next:    "1GpOCgaM",
				HasPrev: true,
				HasNext: false,
			},
		},
		{
			name: "test middle page",
			args: args{
				features: []*Feature{{ID: 3}, {ID: 4}, {ID: 5}, {ID: 6}},
				id: PrevNextID{
					Prev: 2,
					Next: 7,
				},
			},
			want: Cursors{
				Prev:    "GDsXEuZV",
				Next:    "7Temhips",
				HasPrev: true,
				HasNext: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCursors(tt.args.id)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCursors() = %v, want %v", got, tt.want)
			}
		})
	}
}
