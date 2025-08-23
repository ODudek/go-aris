package resp

import "strconv"

type IntValue struct {
	typ string
	num int
}

func (i *IntValue) GetValue() []byte {
	numStr := strconv.Itoa(i.num)
	typLen := len(i.typ)
	numLen := len(numStr)

	bytes := make([]byte, 0, typLen+numLen+2)
	bytes = append(bytes, i.typ...)
	bytes = append(bytes, numStr...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func NewIntValue(num int) *IntValue {
	return &IntValue{
		typ: INTEGER,
		num: num,
	}
}
