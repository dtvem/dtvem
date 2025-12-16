package manifest

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed data/*.json
var embeddedManifests embed.FS

// EmbeddedSource reads manifests from files embedded in the binary.
// Manifest files are stored in the data/ subdirectory.
type EmbeddedSource struct {
	fs fs.FS
}

// NewEmbeddedSource creates a Source that reads from embedded manifest files.
func NewEmbeddedSource() *EmbeddedSource {
	// Get the data subdirectory as the root
	subFS, _ := fs.Sub(embeddedManifests, "data")
	return &EmbeddedSource{fs: subFS}
}

// NewEmbeddedSourceFromFS creates a Source from a custom filesystem.
// This is useful for testing with mock filesystems.
func NewEmbeddedSourceFromFS(fsys fs.FS) *EmbeddedSource {
	return &EmbeddedSource{fs: fsys}
}

// GetManifest reads and parses the manifest for the given runtime.
func (s *EmbeddedSource) GetManifest(runtime string) (*Manifest, error) {
	data, err := fs.ReadFile(s.fs, runtime+".json")
	if err != nil {
		return nil, &ErrManifestNotFound{Runtime: runtime}
	}

	return ParseManifest(data)
}

// ListRuntimes returns all available runtime names by scanning embedded files.
func (s *EmbeddedSource) ListRuntimes() ([]string, error) {
	entries, err := fs.ReadDir(s.fs, ".")
	if err != nil {
		return nil, err
	}

	var runtimes []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			runtimes = append(runtimes, strings.TrimSuffix(name, ".json"))
		}
	}

	return runtimes, nil
}
