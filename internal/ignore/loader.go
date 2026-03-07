package ignore

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Load discovers ignore files under root and returns ordered rules.
func Load(root string, files []string) ([]IgnoreRule, error) {
	root = filepath.Clean(root)
	loader := newLoader(root, files)
	if err := loader.walk(); err != nil {
		return nil, err
	}

	return loader.rules, nil
}

type ignoreLoader struct {
	root        string
	files       []string
	directories []string
	rules       []IgnoreRule
}

func newLoader(root string, files []string) *ignoreLoader {
	return &ignoreLoader{
		root:  root,
		files: append([]string{}, files...),
	}
}

func (loader *ignoreLoader) walk() error {
	if err := filepath.WalkDir(loader.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		loader.directories = append(loader.directories, path)

		return nil
	}); err != nil {
		return err
	}

	sort.Strings(loader.directories)

	for _, directory := range loader.directories {
		relBase, err := filepath.Rel(loader.root, directory)
		if err != nil {
			return err
		}
		relBase = filepath.ToSlash(relBase)
		if relBase == "." {
			relBase = ""
		}
		if err := loader.loadDirectory(directory, relBase); err != nil {
			return err
		}
	}

	return nil
}

func (loader *ignoreLoader) loadDirectory(directory string, relBase string) error {
	baseDir := relBase
	if baseDir == "" {
		baseDir = "."
	}
	for _, fileName := range loader.files {
		fullPath := filepath.Join(directory, fileName)
		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return err
		}
		if info.IsDir() {
			continue
		}
		source := fileName
		if relBase != "" {
			source = filepath.ToSlash(filepath.Join(relBase, fileName))
		}
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		rules, err := Parse(baseDir, source, string(content))
		if err != nil {
			return err
		}
		loader.rules = append(loader.rules, rules...)
	}

	return nil
}
