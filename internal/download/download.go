package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func LocalSourceFile(path string) (SourceFile, error) {
	sum, size, err := SHA256File(path)
	if err != nil {
		return SourceFile{}, err
	}
	return SourceFile{Name: filepath.Base(path), URL: "file://" + path, LocalPath: path, SHA256: sum, SizeBytes: size}, nil
}

func DownloadAll(ctx context.Context, cacheDir, group string, urls []string) ([]SourceFile, error) {
	var out []SourceFile
	if len(urls) == 0 {
		return out, nil
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	for _, raw := range urls {
		sf, err := Download(ctx, cacheDir, group, raw)
		if err != nil {
			return nil, err
		}
		out = append(out, sf)
	}
	return out, nil
}

func Download(ctx context.Context, cacheDir, group, raw string) (SourceFile, error) {
	if strings.HasPrefix(raw, "file://") {
		return LocalSourceFile(strings.TrimPrefix(raw, "file://"))
	}
	u, err := url.Parse(raw)
	if err != nil {
		return SourceFile{}, err
	}
	name := filepath.Base(u.Path)
	if name == "." || name == "/" || name == "" {
		name = group + ".dat"
	}
	dir := filepath.Join(cacheDir, group)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return SourceFile{}, err
	}
	local := filepath.Join(dir, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return SourceFile{}, err
	}
	req.Header.Set("User-Agent", "ASNForge/0.1 (https://github.com/ipanalytics/ASNforge)")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return SourceFile{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return SourceFile{}, fmt.Errorf("download %s: HTTP %s", raw, res.Status)
	}
	tmp := local + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return SourceFile{}, err
	}
	_, copyErr := io.Copy(f, res.Body)
	closeErr := f.Close()
	if copyErr != nil {
		return SourceFile{}, copyErr
	}
	if closeErr != nil {
		return SourceFile{}, closeErr
	}
	if err := os.Rename(tmp, local); err != nil {
		return SourceFile{}, err
	}
	sum, size, err := SHA256File(local)
	if err != nil {
		return SourceFile{}, err
	}
	return SourceFile{Name: name, URL: raw, LocalPath: local, SHA256: sum, SizeBytes: size, DownloadedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}
