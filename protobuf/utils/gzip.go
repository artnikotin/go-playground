package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
)

func CompressGZIP(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	var w io.WriteCloser
	var err error
	w, err = gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, nil
	}
	if _, err := w.Write(data); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := w.Close(); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecompressGZIP(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		if err == gzip.ErrHeader || err == io.ErrUnexpectedEOF {
			return data, nil
		}
		return nil, errors.WithStack(err)
	}
	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := gr.Close(); err != nil {
		return nil, errors.WithStack(err)
	}
	return data, nil
}

func GzipAndPrint(body []byte) {
	fmt.Printf("Best speed: %d\n", len(Must2(CompressGZIP(body, gzip.BestSpeed))))
	fmt.Printf("Best compression: %d\n", len(Must2(CompressGZIP(body, gzip.BestCompression))))
	fmt.Printf("Default compression: %d\n", len(Must2(CompressGZIP(body, gzip.DefaultCompression))))
}
