package resp

type IntValue struct {
	typ string
	num int
}

func (i *IntValue) GetValue() []byte {
	return []byte(i.typ + string(rune(i.num)) + "\r\n")
}

func NewIntValue(num int) *IntValue {
	return &IntValue{
		typ: INTEGER,
		num: num,
	}
}
