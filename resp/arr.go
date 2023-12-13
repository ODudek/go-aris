package resp

import "strconv"

type ArrValue struct {
	typ string
	arr []Value
}

func (a *ArrValue) GetValue() []byte {
	l := len(a.arr)
	typLen := len(a.typ)
	numLen := len(strconv.Itoa(l))

	bytes := make([]byte, 0, typLen+numLen+2)
	bytes = append(bytes, a.typ...)
	bytes = strconv.AppendInt(bytes, int64(l), 10)
	bytes = append(bytes, '\r', '\n')

	for _, v := range a.arr {
		valueBytes := v.GetValue()
		bytes = append(bytes, valueBytes...)
	}

	return bytes
}

func NewArrValue(arr []Value) *ArrValue {
	return &ArrValue{
		typ: ARRAY,
		arr: arr,
	}
}
