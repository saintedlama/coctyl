package coctyl

// FileResult holds the structural fingerprint results for a Go file.
type FileResult struct {
	Path string `json:"path,omitempty"`
	Hash string `json:"hash"`
}
