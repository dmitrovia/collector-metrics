package gizp

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"google.golang.org/grpc"
)

const maxCode int = 300

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(
	w http.ResponseWriter,
) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	tp, err := c.zw.Write(p)
	if err != nil {
		return 0, fmt.Errorf("compressWriterWrite: %w", err)
	}

	return tp, nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < maxCode {
		c.w.Header().Set("Content-Encoding", "gzip")
	}

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("compressWriterClose: %w", err)
	}

	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(
	reader io.ReadCloser,
) (*compressReader, error) {
	zipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  reader,
		zr: zipReader,
	}, nil
}

func (c compressReader) Read(p []byte) (int, error) {
	n, err := c.zr.Read(p)

	return n, err
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("compressReaderClose: %w", err)
	}

	err1 := c.zr.Close()
	if err1 != nil {
		return fmt.Errorf("compressReaderClose: %w", err1)
	}

	return nil
}

func DecryptInterceptor(
	params *bizmodels.InitParams,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		var availableCT []string

		defWriter := writer

		accCT := req.Header.Get("Content-Type")

		accCT1 := req.Header.Get("Accept") // for autotests

		acceptEncoding := req.Header.Get("Accept-Encoding")

		availableCT = []string{"application/json", "text/html"}

		supportsGzip := strings.Contains(
			acceptEncoding, "gzip") && (slices.Contains(
			availableCT, accCT) || slices.Contains(
			availableCT, accCT1))
		if supportsGzip {
			cw := newCompressWriter(writer)
			defWriter = cw
			defer cw.Close()
		}

		contentEncoding := req.Header.Get("Content-Encoding")

		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			compressReader, err := newCompressReader(req.Body)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				fmt.Println(" %w",
					err)

				return
			}

			req.Body = compressReader

			defer compressReader.Close()
		}

		return handler(ctx, req)
	}
}
