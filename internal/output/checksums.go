package output

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ipanalytics/ASNforge/internal/download"
)

type Artifact struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	Records     int    `json:"records,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Description string `json:"description,omitempty"`
}

func ArtifactInfo(name, path string, records int, contentType, desc string) (Artifact, error) {
	sum, size, err := download.SHA256File(path)
	if err != nil {
		return Artifact{}, err
	}
	return Artifact{Name: name, Path: path, SHA256: sum, SizeBytes: size, Records: records, ContentType: contentType, Description: desc}, nil
}

func WriteChecksums(outDir string, artifacts []Artifact) error {
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Name < artifacts[j].Name })
	f, err := os.Create(filepath.Join(outDir, "checksums.txt"))
	if err != nil {
		return err
	}
	defer f.Close()
	for _, a := range artifacts {
		if _, err := fmt.Fprintf(f, "%s  %s\n", a.SHA256, filepath.Base(a.Path)); err != nil {
			return err
		}
	}
	return nil
}
