package resp

import (
	"reflect"
	"testing"
)

func TestArrValue_GetValue(t *testing.T) {
	tests := []struct {
		name   string
		fields ArrValue
		want   []byte
	}{
		{
			name: "array value with one string value",
			fields: ArrValue{
				typ: ARRAY,
				arr: []Value{
					NewStrValue("test"),
				},
			},
			want: NewArrValue([]Value{
				NewStrValue("test"),
			}).GetValue(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.GetValue(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewArrValue(t *testing.T) {
	type args struct {
		arr []Value
	}
	tests := []struct {
		name string
		args args
		want *ArrValue
	}{
		{
			name: "array value with one string value",
			args: args{
				arr: []Value{
					NewStrValue("test"),
				},
			},
			want: &ArrValue{
				typ: ARRAY,
				arr: []Value{
					NewStrValue("test"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewArrValue(tt.args.arr); !reflect.DeepEqual(got.GetValue(), tt.want.GetValue()) {
				t.Errorf("NewArrValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
