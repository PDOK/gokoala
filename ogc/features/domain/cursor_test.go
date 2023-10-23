package domain

import (
	"reflect"
	"testing"
)

func TestNewCursor(t *testing.T) {
	type args struct {
		features []*Feature
		last     bool
	}
	tests := []struct {
		name string
		args args
		want Cursor
	}{
		{
			name: "test first page",
			args: args{
				features: []*Feature{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}},
				last:     false,
			},
			want: Cursor{
				Prev:    "1GpOCgaM",
				Next:    "eVc7GU6Q",
				IsFirst: true,
				IsLast:  false,
			},
		},
		{
			name: "test last page",
			args: args{
				features: []*Feature{{ID: 5}, {ID: 6}, {ID: 7}, {ID: 8}},
				last:     true,
			},
			want: Cursor{
				Prev:    "1GpOCgaM",
				Next:    "VCHYvtZJ",
				IsFirst: false,
				IsLast:  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCursor(tt.args.features, tt.args.last)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}
