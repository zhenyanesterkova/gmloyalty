package mycompress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

const (
	successfulMaxCode = 300
)

type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	n, err := c.zw.Write(p)
	if err != nil {
		return n, fmt.Errorf("gzip.go func Write(): error write - %w", err)
	}
	return n, nil
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < successfulMaxCode {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("gzip.go func Close(): error close - %w", err)
	}
	return nil
}

type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip.go func NewCompressReader(): error create reader - %w", err)
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	n, err = c.zr.Read(p)
	if err != nil {
		return n, fmt.Errorf("gzip.go func Read(): error read - %w", err)
	}
	return n, nil
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("gzip.go func Close(): %w", err)
	}
	if err := c.zr.Close(); err != nil {
		return fmt.Errorf("gzip.go func Close(): %w", err)
	}
	return nil
}
