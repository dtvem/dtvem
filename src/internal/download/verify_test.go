package download

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestComputeSHA256(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create a test file with known content
	content := []byte("hello world\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	// SHA256 of "hello world\n" (with newline)
	expectedHash := "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"

	hash, err := ComputeSHA256(testFile)
	if err != nil {
		t.Fatalf("ComputeSHA256 failed: %v", err)
	}

	if hash != expectedHash {
		t.Errorf("hash = %q, want %q", hash, expectedHash)
	}
}

func TestComputeSHA256FileNotFound(t *testing.T) {
	_, err := ComputeSHA256("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestVerifyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("hello world\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	expectedHash := "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"

	t.Run("valid checksum", func(t *testing.T) {
		err := VerifyFile(testFile, expectedHash)
		if err != nil {
			t.Errorf("VerifyFile failed: %v", err)
		}
	})

	t.Run("valid checksum uppercase", func(t *testing.T) {
		err := VerifyFile(testFile, "A948904F2F0F479B8F8197694B30184B0D2ED1C1CD2A1EC0FB85D299A192A447")
		if err != nil {
			t.Errorf("VerifyFile should accept uppercase: %v", err)
		}
	})

	t.Run("valid checksum with whitespace", func(t *testing.T) {
		err := VerifyFile(testFile, "  "+expectedHash+"  ")
		if err != nil {
			t.Errorf("VerifyFile should trim whitespace: %v", err)
		}
	})

	t.Run("invalid checksum", func(t *testing.T) {
		err := VerifyFile(testFile, "0000000000000000000000000000000000000000000000000000000000000000")
		if err == nil {
			t.Error("expected error for invalid checksum")
		}
		var mismatchErr *ErrChecksumMismatch
		if !errors.As(err, &mismatchErr) {
			t.Errorf("expected ErrChecksumMismatch, got %T", err)
			return
		}
		if mismatchErr.Actual != expectedHash {
			t.Errorf("Actual = %q, want %q", mismatchErr.Actual, expectedHash)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		err := VerifyFile("/nonexistent/file.txt", expectedHash)
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestErrChecksumMismatch(t *testing.T) {
	err := &ErrChecksumMismatch{
		Expected: "abc123",
		Actual:   "def456",
	}

	want := "checksum mismatch: expected abc123, got def456"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}
