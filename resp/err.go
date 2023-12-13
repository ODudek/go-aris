package resp

import "strconv"

type ErrValue struct {
	typ string
	str string
}

func (e *ErrValue) GetValue() []byte {
	l := len(e.str)
	typLen := len(e.typ)
	numLen := len(strconv.Itoa(l))

	bytes := make([]byte, 0, typLen+numLen+2)
	bytes = append(bytes, e.typ...)
	bytes = append(bytes, e.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func NewErrValue(str string) *ErrValue {
	return &ErrValue{
		typ: ERROR,
		str: str,
	}
}
