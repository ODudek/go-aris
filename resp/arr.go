package resp

import "strconv"

type ArrValue struct {
	typ string
	arr []Value
}

func (a *ArrValue) GetValue() []byte {
	l := len(a.arr)
	var bytes []byte
	bytes = append(bytes, a.typ...)
	bytes = append(bytes, strconv.Itoa(l)...)
	bytes = append(bytes, '\r', '\n')
	for _, v := range a.arr {
		bytes = append(bytes, v.GetValue()...)
	}
	return bytes
}

func NewArrValue(arr []Value) *ArrValue {
	return &ArrValue{
		typ: ARRAY,
		arr: arr,
	}
}
