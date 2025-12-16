package manifest

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"
)

// mockSource is a test source that tracks calls
type mockSource struct {
	manifests map[string]*Manifest
	callCount map[string]int
	runtimes  []string
	returnErr error
}

func newMockSource() *mockSource {
	return &mockSource{
		manifests: make(map[string]*Manifest),
		callCount: make(map[string]int),
	}
}

func (s *mockSource) GetManifest(runtime string) (*Manifest, error) {
	s.callCount[runtime]++
	if s.returnErr != nil {
		return nil, s.returnErr
	}
	m, ok := s.manifests[runtime]
	if !ok {
		return nil, &ErrManifestNotFound{Runtime: runtime}
	}
	return m, nil
}

func (s *mockSource) ListRuntimes() ([]string, error) {
	return s.runtimes, nil
}

func TestCachedSource(t *testing.T) {
	tmpDir := t.TempDir()

	mock := newMockSource()
	mock.manifests["python"] = &Manifest{
		Version:  1,
		Versions: map[string]map[string]*Download{},
	}
	mock.runtimes = []string{"python"}

	source := NewCachedSource(mock, tmpDir, time.Hour)

	t.Run("first call fetches from source", func(t *testing.T) {
		m, err := source.GetManifest("python")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m == nil {
			t.Fatal("expected manifest, got nil")
		}
		if mock.callCount["python"] != 1 {
			t.Errorf("callCount = %d, want 1", mock.callCount["python"])
		}
	})

	t.Run("second call uses cache", func(t *testing.T) {
		m, err := source.GetManifest("python")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m == nil {
			t.Fatal("expected manifest, got nil")
		}
		// Call count should still be 1 (cached)
		if mock.callCount["python"] != 1 {
			t.Errorf("callCount = %d, want 1 (should use cache)", mock.callCount["python"])
		}
	})

	t.Run("ForceRefresh bypasses cache", func(t *testing.T) {
		m, err := source.ForceRefresh("python")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m == nil {
			t.Fatal("expected manifest, got nil")
		}
		// Call count should be 2 now
		if mock.callCount["python"] != 2 {
			t.Errorf("callCount = %d, want 2", mock.callCount["python"])
		}
	})

	t.Run("ListRuntimes delegates to source", func(t *testing.T) {
		runtimes, err := source.ListRuntimes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(runtimes) != 1 || runtimes[0] != "python" {
			t.Errorf("runtimes = %v, want [python]", runtimes)
		}
	})
}

func TestCachedSourceExpiration(t *testing.T) {
	tmpDir := t.TempDir()

	mock := newMockSource()
	mock.manifests["python"] = &Manifest{
		Version:  1,
		Versions: map[string]map[string]*Download{},
	}

	// Use very short TTL
	source := NewCachedSource(mock, tmpDir, time.Millisecond)

	// First call
	_, err := source.GetManifest("python")
	if err != nil {
		t.Fatal(err)
	}

	// Wait for cache to expire
	time.Sleep(10 * time.Millisecond)

	// Second call should fetch again due to expiration
	_, err = source.GetManifest("python")
	if err != nil {
		t.Fatal(err)
	}

	if mock.callCount["python"] != 2 {
		t.Errorf("callCount = %d, want 2 (cache should have expired)", mock.callCount["python"])
	}
}

func TestCachedSourceClearCache(t *testing.T) {
	tmpDir := t.TempDir()

	mock := newMockSource()
	mock.manifests["python"] = &Manifest{
		Version:  1,
		Versions: map[string]map[string]*Download{},
	}

	source := NewCachedSource(mock, tmpDir, time.Hour)

	// Populate cache
	_, err := source.GetManifest("python")
	if err != nil {
		t.Fatal(err)
	}

	// Verify cache file exists
	cachePath := filepath.Join(tmpDir, "python.cache.json")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file should exist")
	}

	// Clear cache
	if err := source.ClearCache(); err != nil {
		t.Fatalf("ClearCache failed: %v", err)
	}

	// Verify cache file is gone
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Error("cache file should be deleted")
	}

	// Next call should fetch from source again
	_, err = source.GetManifest("python")
	if err != nil {
		t.Fatal(err)
	}

	if mock.callCount["python"] != 2 {
		t.Errorf("callCount = %d, want 2", mock.callCount["python"])
	}
}

func TestCachedSourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mock := newMockSource()

	source := NewCachedSource(mock, tmpDir, time.Hour)

	_, err := source.GetManifest("ruby")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsManifestNotFound(err) {
		t.Errorf("expected ErrManifestNotFound, got %T: %v", err, err)
	}
}

func TestEmbeddedSource(t *testing.T) {
	// Test the real embedded source with the data files
	source := NewEmbeddedSource()

	// We should have python, node, and ruby manifests embedded
	runtimes, err := source.ListRuntimes()
	if err != nil {
		t.Fatalf("ListRuntimes failed: %v", err)
	}

	found := make(map[string]bool)
	for _, r := range runtimes {
		found[r] = true
	}

	if !found["python"] {
		t.Error("expected python in embedded runtimes")
	}
	if !found["node"] {
		t.Error("expected node in embedded runtimes")
	}
	if !found["ruby"] {
		t.Error("expected ruby in embedded runtimes")
	}

	// Test loading a manifest
	m, err := source.GetManifest("python")
	if err != nil {
		t.Fatalf("GetManifest failed: %v", err)
	}
	if m.Version != 1 {
		t.Errorf("Version = %d, want 1", m.Version)
	}
}

func TestNewEmbeddedSourceFromFS(t *testing.T) {
	mockFS := fstest.MapFS{
		"test.json": &fstest.MapFile{
			Data: []byte(`{"version": 1, "versions": {}}`),
		},
	}

	source := NewEmbeddedSourceFromFS(mockFS)

	m, err := source.GetManifest("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Version != 1 {
		t.Errorf("Version = %d, want 1", m.Version)
	}
}
