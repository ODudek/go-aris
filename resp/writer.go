package resp

import "io"

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

func (w *Writer) Write(v Value) error {
	bytes := v.GetValue()
	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
