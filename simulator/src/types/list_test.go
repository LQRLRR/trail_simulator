package types

import (
	"reflect"
	"testing"
)

func TestList_Remove(t *testing.T) {
	type args struct {
		index int
	}
	l := List{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	tests := []struct {
		name string
		l    List
		args args
		want List
	}{
		{
			name: "remove",
			l:    l,
			args: args{5},
			want: List{0, 1, 2, 3, 4, 6, 7, 8, 9},
		},
		{
			name: "remove first element",
			l:    l,
			args: args{0},
			want: List{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.Remove(tt.args.index); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List.Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}
