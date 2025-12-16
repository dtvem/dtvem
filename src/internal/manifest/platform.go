package manifest

import (
	"fmt"
	"runtime"
)

// Platform keys match Go's runtime.GOOS-GOARCH format.
const (
	PlatformWindowsAMD64 = "windows-amd64"
	PlatformWindowsARM64 = "windows-arm64"
	PlatformWindows386   = "windows-386"
	PlatformDarwinAMD64  = "darwin-amd64"
	PlatformDarwinARM64  = "darwin-arm64"
	PlatformLinuxAMD64   = "linux-amd64"
	PlatformLinuxARM64   = "linux-arm64"
	PlatformLinuxARM     = "linux-arm"
	PlatformLinux386     = "linux-386"
)

// CurrentPlatform returns the platform key for the current OS and architecture.
func CurrentPlatform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

// ValidPlatforms returns all supported platform keys.
func ValidPlatforms() []string {
	return []string{
		PlatformWindowsAMD64,
		PlatformWindowsARM64,
		PlatformWindows386,
		PlatformDarwinAMD64,
		PlatformDarwinARM64,
		PlatformLinuxAMD64,
		PlatformLinuxARM64,
		PlatformLinuxARM,
		PlatformLinux386,
	}
}

// IsValidPlatform checks if a platform key is valid.
func IsValidPlatform(platform string) bool {
	for _, p := range ValidPlatforms() {
		if p == platform {
			return true
		}
	}
	return false
}
