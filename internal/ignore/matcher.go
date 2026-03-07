package ignore

import (
	"fmt"
	"path"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Matcher evaluates ignore rules against root-relative paths.
type Matcher struct {
	rules []IgnoreRule
}

// NewMatcher returns an immutable matcher for the provided rules.
func NewMatcher(rules []IgnoreRule) *Matcher {
	cloned := append([]IgnoreRule{}, rules...)

	return &Matcher{rules: cloned}
}

// Match reports whether relPath should be ignored.
func Match(rules []IgnoreRule, relPath string, isDir bool) (bool, error) {
	matcher := NewMatcher(rules)

	return matcher.Ignored(relPath, isDir)
}

// Ignored reports whether relPath should be ignored.
func (matcher *Matcher) Ignored(relPath string, isDir bool) (bool, error) {
	cleaned := path.Clean(strings.TrimPrefix(relPath, "./"))
	if cleaned == "." {
		return false, nil
	}

	ignored := false
	for _, rule := range matcher.rules {
		matched, err := ruleMatchesPath(rule, cleaned, isDir)
		if err != nil {
			return false, err
		}
		if matched {
			ignored = !rule.Negated
		}
	}

	return ignored, nil
}

func ruleMatchesPath(rule IgnoreRule, relPath string, isDir bool) (bool, error) {
	if rule.DirectoryOnly && !isDir {
		return matchParentDirectory(rule, relPath)
	}

	return matchRule(rule, relPath)
}

func matchParentDirectory(rule IgnoreRule, relPath string) (bool, error) {
	dirPath := path.Dir(relPath)
	if dirPath == "." || dirPath == "/" {
		return false, nil
	}

	segments := strings.Split(dirPath, "/")
	for i := 1; i <= len(segments); i++ {
		candidate := strings.Join(segments[:i], "/")
		matched, err := matchRule(rule, candidate)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

func matchRule(rule IgnoreRule, relPath string) (bool, error) {
	rel, ok := trimBaseDir(rule.BaseDir, relPath)
	if !ok {
		return false, nil
	}
	if rel == "" {
		return false, nil
	}

	pattern := rule.Pattern
	if strings.HasPrefix(pattern, "/") {
		pattern = strings.TrimPrefix(pattern, "/")

		return matchPattern(rule, pattern, rel)
	}
	if strings.Contains(pattern, "/") {
		return matchPattern(rule, pattern, rel)
	}

	joinedPattern := "**/" + pattern

	return matchPattern(rule, joinedPattern, rel)
}

func trimBaseDir(baseDir string, relPath string) (string, bool) {
	if baseDir == "" || baseDir == "." {
		return relPath, true
	}

	baseDir = strings.TrimPrefix(baseDir, "./")
	if relPath == baseDir {
		return "", true
	}
	if strings.HasPrefix(relPath, baseDir+"/") {
		return strings.TrimPrefix(relPath, baseDir+"/"), true
	}

	return "", false
}

func matchPattern(rule IgnoreRule, pattern string, relPath string) (bool, error) {
	matched, err := doublestar.Match(pattern, relPath)
	if err != nil {
		return false, fmt.Errorf("%s:%d: invalid ignore pattern", rule.Source, rule.Line)
	}

	return matched, nil
}
