package types

import (
	"reflect"
	"testing"
)

func TestUint256_Max(t *testing.T) {
	var max [32]byte
	for i := 0; i < 32; i++ {
		max[i] = uint8(255)
	}
	tests := []struct {
		name string
		u    Uint256
		want Uint256
	}{
		{
			name: "max",
			u:    Uint256{},
			want: max,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Max(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint256.Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint256_AddUint8(t *testing.T) {
	type args struct {
		a int
	}

	tests := []struct {
		name string
		u    Uint256
		args args
		want Uint256
	}{
		{
			name: "add 0",
			u:    Uint256{},
			args: args{0},
			want: Uint256{},
		},
		{
			name: "add 1",
			u:    Uint256{},
			args: args{1},
			want: Uint256{1},
		},
		{
			name: "add 1 to 255",
			u:    Uint256{255},
			args: args{1},
			want: Uint256{0, 1},
		},
		{
			name: "add 1 to max",
			u:    Uint256{}.Max(),
			args: args{1},
			want: Uint256{0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.AddUint8(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint256.AddUint8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint256_Add(t *testing.T) {
	type args struct {
		a Uint256
	}
	tests := []struct {
		name string
		u    Uint256
		args args
		want Uint256
	}{
		{
			name: "add 0",
			u:    Uint256{},
			args: args{Uint256{}},
			want: Uint256{},
		},
		{
			name: "add 1",
			u:    Uint256{},
			args: args{Uint256{1}},
			want: Uint256{1},
		},
		{
			name: "add 1 to 255",
			u:    Uint256{255},
			args: args{Uint256{1}},
			want: Uint256{0, 1},
		},
		{
			name: "add 1 to max",
			u:    Uint256{}.Max(),
			args: args{Uint256{1}},
			want: Uint256{0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Add(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint256.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint256_Divide2(t *testing.T) {
	tests := []struct {
		name string
		u    Uint256
		want Uint256
	}{
		{
			name: "devide 1 by 2",
			u:    Uint256{1},
			want: Uint256{},
		},
		{
			name: "devide 2 by 2",
			u:    Uint256{2},
			want: Uint256{1},
		},
		{
			name: "devide 256 by 2",
			u:    Uint256{0, 1},
			want: Uint256{128},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Divide2(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint256.Divide2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint256_Larger(t *testing.T) {
	type args struct {
		b Uint256
	}
	tests := []struct {
		name string
		u    Uint256
		args args
		want bool
	}{
		{
			name: "0 not larger than 0",
			u:    Uint256{},
			args: args{Uint256{}},
			want: false,
		},
		{
			name: "1 larger than 0",
			u:    Uint256{1},
			args: args{Uint256{}},
			want: true,
		},
		{
			name: "1 not larger than 2",
			u:    Uint256{1},
			args: args{Uint256{2}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Larger(tt.args.b); got != tt.want {
				t.Errorf("Uint256.Larger() = %v, want %v", got, tt.want)
			}
		})
	}
}
