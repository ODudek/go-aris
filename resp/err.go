package resp

type ErrValue struct {
	typ string
	str string
}

func (e *ErrValue) GetValue() []byte {
	var bytes []byte
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
