package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	baselineSchemaVersion = 1

	windowsDrivePrefixLen  = 3
	windowsDriveLetterIdx  = 0
	windowsDriveSeparator  = 1
	windowsAbsolutePathIdx = 2
)

type baselineDocumentRaw struct {
	SchemaVersion *int     `json:"schemaVersion"`
	Entries       *[]Entry `json:"entries"`
}

// Load reads and validates a baseline JSON file.
func Load(path string) (Document, error) {
	data, err := os.ReadFile(path) //#nosec G304 -- caller controls path and validation happens on parsed content
	if err != nil {
		return Document{}, fmt.Errorf("read baseline: %w", err)
	}

	var raw baselineDocumentRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return Document{}, fmt.Errorf("parse baseline: %w", err)
	}

	if err := validateRawDocument(raw); err != nil {
		return Document{}, err
	}

	entries := append([]Entry{}, (*raw.Entries)...)
	document := Document{
		SchemaVersion: *raw.SchemaVersion,
		Entries:       entries,
	}

	if err := validateEntries(document.Entries); err != nil {
		return Document{}, err
	}

	return document, nil
}

func validateRawDocument(raw baselineDocumentRaw) error {
	if raw.SchemaVersion == nil {
		return fmt.Errorf("baseline schemaVersion is required")
	}
	if *raw.SchemaVersion != baselineSchemaVersion {
		return fmt.Errorf("baseline schemaVersion must be %d", baselineSchemaVersion)
	}
	if raw.Entries == nil {
		return fmt.Errorf("baseline entries is required")
	}

	return nil
}

func validateEntries(entries []Entry) error {
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		normalizedPath, err := validateEntryFilePath(entry.FilePath)
		if err != nil {
			return err
		}
		if strings.TrimSpace(entry.Message) == "" {
			return fmt.Errorf("baseline entry message is required")
		}
		if entry.Count <= 0 {
			return fmt.Errorf("baseline entry count must be positive")
		}

		key := normalizedPath + "\x00" + entry.Message
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate baseline entry for filePath=%q message=%q", normalizedPath, entry.Message)
		}
		seen[key] = struct{}{}
	}

	return nil
}

func validateEntryFilePath(filePath string) (string, error) {
	trimmed := strings.TrimSpace(filePath)
	if trimmed == "" {
		return "", fmt.Errorf("baseline entry filePath is required")
	}

	normalized := path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))
	if strings.HasPrefix(normalized, "/") ||
		filepath.IsAbs(trimmed) ||
		isWindowsAbsolutePath(normalized) ||
		normalized == "." ||
		normalized == ".." ||
		strings.HasPrefix(normalized, "../") {
		return "", fmt.Errorf("baseline entry filePath must be a normalized relative path")
	}
	if trimmed != normalized {
		return "", fmt.Errorf("baseline entry filePath must be a normalized relative path")
	}

	return normalized, nil
}

func isWindowsAbsolutePath(value string) bool {
	if len(value) < windowsDrivePrefixLen {
		return false
	}

	drive := value[windowsDriveLetterIdx]
	if (drive < 'A' || drive > 'Z') && (drive < 'a' || drive > 'z') {
		return false
	}

	return value[windowsDriveSeparator] == ':' && value[windowsAbsolutePathIdx] == '/'
}
