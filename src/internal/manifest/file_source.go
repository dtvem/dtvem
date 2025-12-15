package manifest

import (
	"os"
	"path/filepath"
	"strings"
)

// FileSource reads manifests from a directory on the filesystem.
// Each runtime has a JSON file named "<runtime>.json" in the directory.
type FileSource struct {
	dir string
}

// NewFileSource creates a Source that reads manifests from the given directory.
func NewFileSource(dir string) *FileSource {
	return &FileSource{dir: dir}
}

// GetManifest reads and parses the manifest for the given runtime.
func (s *FileSource) GetManifest(runtime string) (*Manifest, error) {
	path := filepath.Join(s.dir, runtime+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ErrManifestNotFound{Runtime: runtime}
		}
		return nil, err
	}

	return ParseManifest(data)
}

// ListRuntimes returns all available runtime names by scanning for .json files.
func (s *FileSource) ListRuntimes() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
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
