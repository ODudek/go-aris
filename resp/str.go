package resp

type StrValue struct {
	typ string
	str string
}

func (s *StrValue) GetValue() []byte {
	typLen := len(s.typ)
	strLen := len(s.str)

	bytes := make([]byte, 0, typLen+strLen+2)
	bytes = append(bytes, s.typ...)
	bytes = append(bytes, s.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func NewStrValue(str string) *StrValue {
	return &StrValue{
		typ: STRING,
		str: str,
	}
}
