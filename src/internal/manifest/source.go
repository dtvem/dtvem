package manifest

import (
	"errors"
	"fmt"
)

// Source is the interface for retrieving manifests from various backends.
// Implementations include embedded files, filesystem, and remote HTTP.
type Source interface {
	// GetManifest retrieves the manifest for a runtime (e.g., "python", "node").
	// Returns an error if the manifest cannot be loaded or parsed.
	GetManifest(runtime string) (*Manifest, error)

	// ListRuntimes returns all available runtime names.
	ListRuntimes() ([]string, error)
}

// ErrManifestNotFound is returned when a manifest for a runtime doesn't exist.
type ErrManifestNotFound struct {
	Runtime string
}

func (e *ErrManifestNotFound) Error() string {
	return fmt.Sprintf("manifest not found for runtime: %s", e.Runtime)
}

// IsManifestNotFound checks if an error indicates a missing manifest.
func IsManifestNotFound(err error) bool {
	var target *ErrManifestNotFound
	return errors.As(err, &target)
}
