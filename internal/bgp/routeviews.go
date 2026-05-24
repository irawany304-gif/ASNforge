package bgp

import (
	"context"
	"fmt"

	"github.com/ipanalytics/ASNforge/internal/download"
)

type PreprocessedSource struct {
	SourceName string
	URLs       []string
	Paths      []string
}

func (s PreprocessedSource) Name() string { return s.SourceName }

func (s PreprocessedSource) Download(ctx context.Context, cacheDir string) ([]download.SourceFile, error) {
	var files []download.SourceFile
	for _, p := range s.Paths {
		sf, err := download.LocalSourceFile(p)
		if err != nil {
			return nil, err
		}
		files = append(files, sf)
	}
	got, err := download.DownloadAll(ctx, cacheDir, s.SourceName, s.URLs)
	if err != nil {
		return nil, err
	}
	files = append(files, got...)
	return files, nil
}

func (s PreprocessedSource) Parse(ctx context.Context, files []download.SourceFile) ([]PrefixOriginObservation, error) {
	var out []PrefixOriginObservation
	for _, f := range files {
		obs, err := ParsePreprocessedFile(f.LocalPath)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", f.LocalPath, err)
		}
		out = append(out, obs...)
	}
	return out, nil
}
