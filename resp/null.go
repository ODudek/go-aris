package resp

type NullValue struct {
	typ string
}

func (n *NullValue) GetValue() []byte {
	var bytes []byte
	bytes = append(bytes, n.typ...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func NewNullValue() *NullValue {
	return &NullValue{
		typ: NULL,
	}
}
