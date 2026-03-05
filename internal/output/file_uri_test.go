//nolint:testpackage
package output

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileURIWithLineEncodesSpaces(t *testing.T) {
	t.Parallel()

	filePath := "path with space/file.go"
	fileURI, err := fileURIWithLine(filePath, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(fileURI, "%20") {
		t.Fatalf("expected url encoding for spaces, got %s", fileURI)
	}

	assertFileURIEndsWithLine(t, fileURI, 7)
	assertFileURIPathMatches(t, fileURI, filePath, 7)
}

func TestIsWindowsDrivePath(t *testing.T) {
	t.Parallel()

	if !isWindowsDrivePath("C:/path/to/file.go") {
		t.Fatalf("expected windows drive path")
	}
	if !isWindowsDrivePath("z:/") {
		t.Fatalf("expected windows drive path")
	}
	if isWindowsDrivePath("/C:/path/to/file.go") {
		t.Fatalf("did not expect windows drive path")
	}
	if isWindowsDrivePath("1:/path") {
		t.Fatalf("did not expect windows drive path")
	}
	if isWindowsDrivePath(":/") {
		t.Fatalf("did not expect windows drive path")
	}
}

func TestFileURIWithLineWindowsDrivePrefix(t *testing.T) {
	t.Parallel()

	if filepath.VolumeName("C:/path/to/file.go") == "" {
		t.Skip("volume name detection unavailable on this platform")
	}

	fileURI, err := fileURIWithLine("C:/path/to/file.go", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	uri, err := url.Parse(fileURI)
	if err != nil {
		t.Fatalf("unexpected uri parse error: %v", err)
	}

	if !strings.HasPrefix(uri.Path, "/C:") {
		t.Fatalf("expected windows drive prefix, got %s", uri.Path)
	}
}

func TestFileURIWithLineReturnsErrorWhenCwdMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("current directory removal is restricted on windows")
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected getwd error: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("unexpected chdir restore error: %v", err)
		}
	}()

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("unexpected chdir error: %v", err)
	}
	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("unexpected remove error: %v", err)
	}

	_, err = fileURIWithLine("relative/file.go", 1)
	if err == nil {
		t.Fatalf("expected error with missing cwd")
	}
}

func assertFileURIEndsWithLine(t *testing.T, fileURI string, line int) {
	t.Helper()

	expected := fmt.Sprintf(":%d", line)
	if !strings.HasSuffix(fileURI, expected) {
		t.Fatalf("expected line suffix %s, got %s", expected, fileURI)
	}
}

func assertFileURIPathMatches(t *testing.T, fileURI string, filePath string, line int) {
	t.Helper()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		t.Fatalf("unexpected abs path error: %v", err)
	}

	uri, err := url.Parse(fileURI)
	if err != nil {
		t.Fatalf("unexpected uri parse error: %v", err)
	}

	if uri.Scheme != "file" {
		t.Fatalf("unexpected scheme: %s", uri.Scheme)
	}

	expectedPath := filepath.ToSlash(absPath)
	if isWindowsDrivePath(expectedPath) && !strings.HasPrefix(expectedPath, "/") {
		expectedPath = "/" + expectedPath
	}

	lineSuffix := fmt.Sprintf(":%d", line)
	if !strings.HasSuffix(uri.Path, lineSuffix) {
		t.Fatalf("unexpected uri path suffix: %s", uri.Path)
	}

	pathWithoutLine := strings.TrimSuffix(uri.Path, lineSuffix)
	if expectedPath != pathWithoutLine {
		t.Fatalf("unexpected uri path: %s", uri.Path)
	}
}
