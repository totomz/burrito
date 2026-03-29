package common

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
)

func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		slog.Error("can't write gzip data", "error", err)
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		slog.Error("can't close gzip writer", "error", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func GzipDecompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		slog.Error("can't create gzip reader", "error", err)
		return nil, err
	}
	defer func(reader *gzip.Reader) {
		err := reader.Close()
		if err != nil {
			slog.Error("can't close gzip reader", "error", err)
		}
	}(reader)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		slog.Error("can't read gzip data", "error", err)
		return nil, err
	}
	return buf.Bytes(), nil
}
