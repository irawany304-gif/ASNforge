package download

type SourceFile struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	LocalPath       string `json:"local_path"`
	SHA256          string `json:"sha256"`
	SizeBytes       int64  `json:"size_bytes"`
	DownloadedAt    string `json:"downloaded_at"`
	SourceTimestamp string `json:"source_timestamp,omitempty"`
}
