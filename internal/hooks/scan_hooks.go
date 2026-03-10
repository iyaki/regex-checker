// Package hooks defines optional scan lifecycle hook contracts.
package hooks

import "sort"

// RunContext carries shared run-level data for hook execution.
type RunContext struct {
	Mode             string
	DiffTarget       string
	WorkingDir       string
	AddedLinesOnly   bool
	GitignoreEnabled bool
	AddedLinesByFile map[string]map[int]struct{}
}

// MatchContext carries the minimum match fields needed by post-match hooks.
type MatchContext struct {
	FilePath string
	Line     int
}

// CandidateScope carries candidate-file and optional line-scope constraints.
type CandidateScope struct {
	CandidateFiles   []string
	AddedLinesByFile map[string]map[int]struct{}
}

// IgnoreAugmentation carries additional ignore-file names to load.
type IgnoreAugmentation struct {
	Files []string
}

// Provider defines optional scan lifecycle hooks.
type Provider interface {
	OnCapabilitiesCheck(ctx RunContext) error
	BeforeCollectCandidates(ctx RunContext) (CandidateScope, error)
	BeforeIgnoreEvaluation(ctx RunContext) (IgnoreAugmentation, error)
	AfterMatch(ctx RunContext, match MatchContext) (bool, error)
}

// Registry executes hook providers in deterministic order.
type Registry struct {
	providers []Provider
}

// NewRegistry creates a new registry from providers in call-site order.
func NewRegistry(providers ...Provider) Registry {
	filtered := make([]Provider, 0, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		filtered = append(filtered, provider)
	}

	return Registry{providers: filtered}
}

// Enabled reports whether any hook provider is registered.
func (registry Registry) Enabled() bool {
	return len(registry.providers) > 0
}

// OnCapabilitiesCheck runs capability checks in provider order.
func (registry Registry) OnCapabilitiesCheck(ctx RunContext) error {
	for _, provider := range registry.providers {
		if err := provider.OnCapabilitiesCheck(ctx); err != nil {
			return err
		}
	}

	return nil
}

// BeforeCollectCandidates merges candidate scopes from all providers.
func (registry Registry) BeforeCollectCandidates(ctx RunContext) (CandidateScope, error) {
	files := make([]string, 0)
	fileSeen := make(map[string]struct{})
	var addedLinesByFile map[string]map[int]struct{}

	for _, provider := range registry.providers {
		scope, err := provider.BeforeCollectCandidates(ctx)
		if err != nil {
			return CandidateScope{}, err
		}

		for _, candidateFile := range scope.CandidateFiles {
			if candidateFile == "" {
				continue
			}
			if _, exists := fileSeen[candidateFile]; exists {
				continue
			}

			fileSeen[candidateFile] = struct{}{}
			files = append(files, candidateFile)
		}

		addedLinesByFile = mergeAddedLinesByFile(addedLinesByFile, scope.AddedLinesByFile)
	}

	sort.Strings(files)

	return CandidateScope{CandidateFiles: files, AddedLinesByFile: addedLinesByFile}, nil
}

// BeforeIgnoreEvaluation merges ignore-file augmentations in provider order.
func (registry Registry) BeforeIgnoreEvaluation(ctx RunContext) (IgnoreAugmentation, error) {
	files := make([]string, 0)
	seen := make(map[string]struct{})

	for _, provider := range registry.providers {
		augmentation, err := provider.BeforeIgnoreEvaluation(ctx)
		if err != nil {
			return IgnoreAugmentation{}, err
		}
		for _, fileName := range augmentation.Files {
			if fileName == "" {
				continue
			}
			if _, exists := seen[fileName]; exists {
				continue
			}

			seen[fileName] = struct{}{}
			files = append(files, fileName)
		}
	}

	return IgnoreAugmentation{Files: files}, nil
}

// AfterMatch runs post-match filters in provider order.
func (registry Registry) AfterMatch(ctx RunContext, match MatchContext) (bool, error) {
	for _, provider := range registry.providers {
		keep, err := provider.AfterMatch(ctx, match)
		if err != nil {
			return false, err
		}
		if !keep {
			return false, nil
		}
	}

	return true, nil
}

func mergeAddedLinesByFile(
	target map[string]map[int]struct{},
	next map[string]map[int]struct{},
) map[string]map[int]struct{} {
	if len(next) == 0 {
		return target
	}
	if target == nil {
		target = make(map[string]map[int]struct{}, len(next))
	}

	for filePath, lineSet := range next {
		if len(lineSet) == 0 {
			continue
		}

		targetLineSet, exists := target[filePath]
		if !exists {
			targetLineSet = make(map[int]struct{}, len(lineSet))
			target[filePath] = targetLineSet
		}

		for line := range lineSet {
			targetLineSet[line] = struct{}{}
		}
	}

	return target
}
