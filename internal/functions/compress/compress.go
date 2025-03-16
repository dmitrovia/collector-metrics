// Package compress provides functions
// working with compression.
package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// DeflateCompress - produces deflate compression.
func DeflateCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	writer := gzip.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, fmt.Errorf("DeflateCompress->Write: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	return buf.Bytes(), nil
}

// DeflateDecompress - produces deflate decompression.
func DeflateDecompress(data io.Reader) ([]byte, error) {
	reader, err := gzip.NewReader(data)
	if err != nil {
		defer reader.Close()

		return nil, fmt.Errorf("DeflateDecompress->NewR: %w", err)
	}

	defer reader.Close()

	var buf bytes.Buffer

	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("DeflateDecompress->Read: %w", err)
	}

	return buf.Bytes(), nil
}
