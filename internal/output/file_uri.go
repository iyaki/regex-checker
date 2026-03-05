package output

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func fileURIWithLine(filePath string, line int) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	normalized := filepath.ToSlash(absPath)
	if isWindowsDrivePath(normalized) && !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	uri := url.URL{Scheme: "file", Path: normalized}

	return fmt.Sprintf("%s:%d", uri.String(), line), nil
}

const windowsDrivePrefixLength = 2

func isWindowsDrivePath(path string) bool {
	if len(path) < windowsDrivePrefixLength {
		return false
	}
	letter := path[0]
	if !(letter >= 'A' && letter <= 'Z' || letter >= 'a' && letter <= 'z') {
		return false
	}

	return path[1] == ':'
}
