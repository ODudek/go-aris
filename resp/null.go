package resp

type NullValue struct {
	typ string
}

func (n *NullValue) GetValue() []byte {
	typLen := len(n.typ)

	bytes := make([]byte, 0, typLen+2)
	bytes = append(bytes, n.typ...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func NewNullValue() *NullValue {
	return &NullValue{
		typ: NULL,
	}
}
