package resp

import "strconv"

type BulkValue struct {
	typ  string
	bulk []byte
}

func (b *BulkValue) GetValue() []byte {
	var bytes []byte
	bytes = append(bytes, b.typ...)
	bytes = append(bytes, strconv.Itoa(len(b.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, b.bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func NewBulkValue(bulk []byte) *BulkValue {
	return &BulkValue{
		typ:  BULK,
		bulk: bulk,
	}
}
