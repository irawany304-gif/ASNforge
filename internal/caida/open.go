package caida

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"strings"
)

type readCloser struct {
	io.Reader
	close func() error
}

func (r readCloser) Close() error { return r.close() }

func openMaybeCompressed(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	switch {
	case strings.HasSuffix(path, ".gz"):
		gz, err := gzip.NewReader(f)
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		return readCloser{Reader: gz, close: func() error {
			err1 := gz.Close()
			err2 := f.Close()
			if err1 != nil {
				return err1
			}
			return err2
		}}, nil
	case strings.HasSuffix(path, ".bz2"):
		return readCloser{Reader: bzip2.NewReader(f), close: f.Close}, nil
	default:
		return f, nil
	}
}
