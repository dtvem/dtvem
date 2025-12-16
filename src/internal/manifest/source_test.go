package manifest

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestFileSource(t *testing.T) {
	// Create temp directory with test manifests
	tmpDir := t.TempDir()

	pythonManifest := `{
		"version": 1,
		"versions": {
			"3.13.1": {
				"windows-amd64": {"url": "https://example.com/python.zip", "sha256": "abc123"}
			}
		}
	}`

	nodeManifest := `{
		"version": 1,
		"versions": {
			"22.0.0": {
				"windows-amd64": {"url": "https://example.com/node.zip", "sha256": "def456"}
			}
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "python.json"), []byte(pythonManifest), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "node.json"), []byte(nodeManifest), 0644); err != nil {
		t.Fatal(err)
	}

	source := NewFileSource(tmpDir)

	t.Run("GetManifest existing", func(t *testing.T) {
		m, err := source.GetManifest("python")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m == nil {
			t.Fatal("expected manifest, got nil")
		}
		if len(m.Versions) != 1 {
			t.Errorf("len(Versions) = %d, want 1", len(m.Versions))
		}
	})

	t.Run("GetManifest not found", func(t *testing.T) {
		_, err := source.GetManifest("ruby")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsManifestNotFound(err) {
			t.Errorf("expected ErrManifestNotFound, got %T: %v", err, err)
		}
	})

	t.Run("ListRuntimes", func(t *testing.T) {
		runtimes, err := source.ListRuntimes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(runtimes) != 2 {
			t.Errorf("len(runtimes) = %d, want 2; got %v", len(runtimes), runtimes)
		}

		found := make(map[string]bool)
		for _, r := range runtimes {
			found[r] = true
		}
		if !found["python"] || !found["node"] {
			t.Errorf("expected python and node, got %v", runtimes)
		}
	})
}

func TestEmbeddedSourceFromFS(t *testing.T) {
	// Create a mock filesystem
	mockFS := fstest.MapFS{
		"python.json": &fstest.MapFile{
			Data: []byte(`{
				"version": 1,
				"versions": {
					"3.13.1": {
						"windows-amd64": {"url": "https://example.com/python.zip", "sha256": "abc123"}
					}
				}
			}`),
		},
		"node.json": &fstest.MapFile{
			Data: []byte(`{
				"version": 1,
				"versions": {
					"22.0.0": {
						"linux-amd64": {"url": "https://example.com/node.tar.gz", "sha256": "def456"}
					}
				}
			}`),
		},
	}

	source := NewEmbeddedSourceFromFS(mockFS)

	t.Run("GetManifest existing", func(t *testing.T) {
		m, err := source.GetManifest("python")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m == nil {
			t.Fatal("expected manifest, got nil")
		}
		d := m.GetDownload("3.13.1", "windows-amd64")
		if d == nil {
			t.Fatal("expected download info")
		}
		if d.URL != "https://example.com/python.zip" {
			t.Errorf("URL = %q, want %q", d.URL, "https://example.com/python.zip")
		}
	})

	t.Run("GetManifest not found", func(t *testing.T) {
		_, err := source.GetManifest("ruby")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsManifestNotFound(err) {
			t.Errorf("expected ErrManifestNotFound, got %T: %v", err, err)
		}
	})

	t.Run("ListRuntimes", func(t *testing.T) {
		runtimes, err := source.ListRuntimes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(runtimes) != 2 {
			t.Errorf("len(runtimes) = %d, want 2", len(runtimes))
		}
	})
}

func TestErrManifestNotFound(t *testing.T) {
	err := &ErrManifestNotFound{Runtime: "ruby"}

	if err.Error() != "manifest not found for runtime: ruby" {
		t.Errorf("Error() = %q, want %q", err.Error(), "manifest not found for runtime: ruby")
	}

	if !IsManifestNotFound(err) {
		t.Error("IsManifestNotFound should return true")
	}

	otherErr := os.ErrNotExist
	if IsManifestNotFound(otherErr) {
		t.Error("IsManifestNotFound should return false for other errors")
	}
}
