package manifest

import (
	"runtime"
	"testing"
)

func TestCurrentPlatform(t *testing.T) {
	got := CurrentPlatform()
	want := runtime.GOOS + "-" + runtime.GOARCH

	if got != want {
		t.Errorf("CurrentPlatform() = %q, want %q", got, want)
	}
}

func TestValidPlatforms(t *testing.T) {
	platforms := ValidPlatforms()

	// Should have at least the main platforms
	expected := map[string]bool{
		"windows-amd64": true,
		"darwin-amd64":  true,
		"darwin-arm64":  true,
		"linux-amd64":   true,
	}

	for p := range expected {
		found := false
		for _, vp := range platforms {
			if vp == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidPlatforms() missing %q", p)
		}
	}
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		platform string
		want     bool
	}{
		{"windows-amd64", true},
		{"darwin-arm64", true},
		{"linux-amd64", true},
		{"linux-386", true},
		{"invalid", false},
		{"", false},
		{"windows", false},
		{"amd64", false},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			got := IsValidPlatform(tt.platform)
			if got != tt.want {
				t.Errorf("IsValidPlatform(%q) = %v, want %v", tt.platform, got, tt.want)
			}
		})
	}
}
