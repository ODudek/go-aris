package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(reader),
	}
}

func (r *Reader) Read() (Value, error) {
	t, err := r.reader.ReadByte()
	if err != nil {
		return nil, err
	}
	_type := string(t)
	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unkonwn type: %s\n", _type)
		return nil, nil
	}
}

func (r *Reader) readArray() (Value, error) {
	l, _, err := r.readInteger()
	if err != nil {
		return nil, err
	}
	a := make([]Value, l)
	for i := 0; i < l; i++ {
		val, err := r.Read()
		if err != nil {
			return NewArrValue(a), err
		}
		a[i] = val
	}
	return NewArrValue(a), nil
}

func (r *Reader) readBulk() (Value, error) {
	l, _, err := r.readInteger()
	if err != nil {
		return nil, err
	}
	b := make([]byte, l)
	r.reader.Read(b)
	v := NewBulkValue(b)
	r.readLine()
	return v, nil
}

func (r *Reader) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, n, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *Reader) readLine() (line []byte, numOfChars int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, numOfChars, err
		}
		numOfChars += 1
		line = append(line, b)
		if numOfChars >= 2 && line[numOfChars-2] == '\r' {
			break
		}
	}
	return line[:numOfChars-2], numOfChars, nil
}
