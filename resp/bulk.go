package resp

import "strconv"

type BulkValue struct {
	typ  string
	bulk []byte
}

func (b *BulkValue) GetValue() []byte {
	typLen := len(b.typ)
	bulkLen := len(b.bulk)

	bytes := make([]byte, 0, typLen+len(strconv.Itoa(bulkLen))+4)
	bytes = append(bytes, b.typ...)
	bytes = strconv.AppendInt(bytes, int64(bulkLen), 10)
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
