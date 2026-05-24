package output

import (
	"compress/gzip"
	"io"
	"os"
)

func GzipFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(path + ".gz")
	if err != nil {
		return err
	}
	defer out.Close()
	gw, err := gzip.NewWriterLevel(out, gzip.BestCompression)
	if err != nil {
		return err
	}
	if _, err := io.Copy(gw, in); err != nil {
		_ = gw.Close()
		return err
	}
	return gw.Close()
}
