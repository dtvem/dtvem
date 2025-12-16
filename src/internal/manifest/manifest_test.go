package manifest

import (
	"sort"
	"testing"
)

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid manifest",
			data: `{
				"version": 1,
				"versions": {
					"3.13.1": {
						"windows-amd64": {"url": "https://example.com/python-3.13.1.zip", "sha256": "abc123"},
						"darwin-amd64": null
					}
				}
			}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `{invalid`,
			wantErr: true,
		},
		{
			name: "unsupported manifest version",
			data: `{
				"version": 2,
				"versions": {}
			}`,
			wantErr: true,
			errMsg:  "unsupported manifest version: 2",
		},
		{
			name: "empty versions",
			data: `{
				"version": 1,
				"versions": {}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := ParseManifest([]byte(tt.data))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("error message = %q, want %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m == nil {
				t.Fatal("expected manifest, got nil")
			}
			if m.Version != 1 {
				t.Errorf("Version = %d, want 1", m.Version)
			}
		})
	}
}

func TestManifestGetDownload(t *testing.T) {
	data := `{
		"version": 1,
		"versions": {
			"3.13.1": {
				"windows-amd64": {"url": "https://example.com/win.zip", "sha256": "abc123"},
				"darwin-amd64": null,
				"linux-amd64": {"url": "https://example.com/linux.tar.gz", "sha256": "def456"}
			},
			"3.12.0": {
				"windows-amd64": {"url": "https://example.com/win312.zip", "sha256": "ghi789"}
			}
		}
	}`

	m, err := ParseManifest([]byte(data))
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	tests := []struct {
		name     string
		version  string
		platform string
		wantURL  string
		wantNil  bool
	}{
		{
			name:     "existing version and platform",
			version:  "3.13.1",
			platform: "windows-amd64",
			wantURL:  "https://example.com/win.zip",
		},
		{
			name:     "null download",
			version:  "3.13.1",
			platform: "darwin-amd64",
			wantNil:  true,
		},
		{
			name:     "missing platform",
			version:  "3.13.1",
			platform: "darwin-arm64",
			wantNil:  true,
		},
		{
			name:     "missing version",
			version:  "3.10.0",
			platform: "windows-amd64",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := m.GetDownload(tt.version, tt.platform)
			if tt.wantNil {
				if d != nil {
					t.Errorf("expected nil, got %+v", d)
				}
				return
			}
			if d == nil {
				t.Fatal("expected download, got nil")
			}
			if d.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", d.URL, tt.wantURL)
			}
		})
	}
}

func TestManifestCheckAvailability(t *testing.T) {
	data := `{
		"version": 1,
		"versions": {
			"3.13.1": {
				"windows-amd64": {"url": "https://example.com/win.zip", "sha256": "abc123"},
				"darwin-amd64": null
			}
		}
	}`

	m, err := ParseManifest([]byte(data))
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	tests := []struct {
		name     string
		version  string
		platform string
		want     Availability
	}{
		{
			name:     "available",
			version:  "3.13.1",
			platform: "windows-amd64",
			want:     AvailabilityAvailable,
		},
		{
			name:     "unavailable (null)",
			version:  "3.13.1",
			platform: "darwin-amd64",
			want:     AvailabilityUnavailable,
		},
		{
			name:     "unknown (missing platform)",
			version:  "3.13.1",
			platform: "linux-arm64",
			want:     AvailabilityUnknown,
		},
		{
			name:     "unknown (missing version)",
			version:  "3.10.0",
			platform: "windows-amd64",
			want:     AvailabilityUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.CheckAvailability(tt.version, tt.platform)
			if got != tt.want {
				t.Errorf("CheckAvailability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManifestListVersions(t *testing.T) {
	data := `{
		"version": 1,
		"versions": {
			"3.13.1": {},
			"3.12.0": {},
			"3.11.5": {}
		}
	}`

	m, err := ParseManifest([]byte(data))
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	versions := m.ListVersions()
	sort.Strings(versions)

	expected := []string{"3.11.5", "3.12.0", "3.13.1"}
	if len(versions) != len(expected) {
		t.Fatalf("len(versions) = %d, want %d", len(versions), len(expected))
	}

	for i, v := range versions {
		if v != expected[i] {
			t.Errorf("versions[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestManifestListAvailableVersions(t *testing.T) {
	data := `{
		"version": 1,
		"versions": {
			"3.13.1": {
				"windows-amd64": {"url": "https://example.com/win.zip", "sha256": "abc123"},
				"linux-amd64": {"url": "https://example.com/linux.tar.gz", "sha256": "def456"}
			},
			"3.12.0": {
				"windows-amd64": {"url": "https://example.com/win312.zip", "sha256": "ghi789"},
				"linux-amd64": null
			},
			"3.11.5": {
				"darwin-arm64": {"url": "https://example.com/mac.tar.gz", "sha256": "jkl012"}
			}
		}
	}`

	m, err := ParseManifest([]byte(data))
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	tests := []struct {
		platform string
		want     []string
	}{
		{
			platform: "windows-amd64",
			want:     []string{"3.12.0", "3.13.1"},
		},
		{
			platform: "linux-amd64",
			want:     []string{"3.13.1"},
		},
		{
			platform: "darwin-arm64",
			want:     []string{"3.11.5"},
		},
		{
			platform: "darwin-amd64",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			got := m.ListAvailableVersions(tt.platform)
			sort.Strings(got)
			sort.Strings(tt.want)

			if len(got) != len(tt.want) {
				t.Fatalf("len(versions) = %d, want %d; got %v", len(got), len(tt.want), got)
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("versions[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}
