package domain

import (
	"reflect"
	"testing"
)

func TestNewCursor(t *testing.T) {
	type args struct {
		features []*Feature
		id       NextPrevID
	}
	var tests = []struct {
		name string
		args args
		want Cursor
	}{
		{
			name: "test first page",
			args: args{
				features: []*Feature{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}},
				id: NextPrevID{
					Prev: 0,
					Next: 4,
				},
			},
			want: Cursor{
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
				id: NextPrevID{
					Prev: 4,
					Next: 0,
				},
			},
			want: Cursor{
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
				id: NextPrevID{
					Prev: 2,
					Next: 7,
				},
			},
			want: Cursor{
				Prev:    "GDsXEuZV",
				Next:    "7Temhips",
				HasPrev: true,
				HasNext: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCursor(tt.args.features, tt.args.id)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}
