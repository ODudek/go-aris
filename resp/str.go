package resp

type StrValue struct {
	typ string
	str string
}

func (s *StrValue) GetValue() []byte {
	var bytes []byte
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
