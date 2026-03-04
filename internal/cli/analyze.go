package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/iyaki/regex-checker/internal/config"
	"github.com/iyaki/regex-checker/internal/output"
	"github.com/iyaki/regex-checker/internal/rules"
	"github.com/iyaki/regex-checker/internal/scan"
)

const (
	defaultConfigPath   = "regex-rules.yaml"
	defaultFormat       = "console"
	defaultMaxFileBytes = int64(5242880)
)

const (
	exitCodeError   = 1
	exitCodeFailOn  = 2
	defaultFileMode = 0o600
)

// Config holds parsed analyze command inputs.
type Config struct {
	ConfigPath       string
	Roots            []string
	Formats          []string
	OutJSON          string
	OutSARIF         string
	Include          []string
	Exclude          []string
	Concurrency      int
	MaxFileSizeBytes int64
	FailOnSeverity   string
}

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)

	return nil
}

// HandleAnalyze executes the analyze command.
func HandleAnalyze(args []string, out *bytes.Buffer) int {
	result, failOn, formats, ruleset, cfg, err := runAnalyze(args)
	if err != nil {
		writeError(out, err)

		return exitCodeError
	}

	if err := renderOutputs(formats, ruleset, cfg, result, out); err != nil {
		writeError(out, err)

		return exitCodeError
	}

	if failOn == "" {
		return 0
	}

	return exitCodeForFailOn(result.Matches, failOn)
}

// ParseAnalyzeArgs parses analyze command arguments into a Config.
func ParseAnalyzeArgs(args []string) (Config, error) {
	var cfg Config

	flagSet := flag.NewFlagSet("analyze", flag.ContinueOnError)
	flagSet.SetOutput(&strings.Builder{})

	flagSet.StringVar(&cfg.ConfigPath, "config", defaultConfigPath, "Path to YAML rules config file.")
	formatValue := flagSet.String("format", defaultFormat, "Comma-separated list of formats.")
	flagSet.StringVar(&cfg.OutJSON, "out-json", "", "Output path for JSON results.")
	flagSet.StringVar(&cfg.OutSARIF, "out-sarif", "", "Output path for SARIF results.")
	var include stringSlice
	var exclude stringSlice
	flagSet.Var(&include, "include", "Repeatable include glob.")
	flagSet.Var(&exclude, "exclude", "Repeatable exclude glob.")
	flagSet.IntVar(&cfg.Concurrency, "concurrency", runtime.GOMAXPROCS(0), "Worker count.")
	flagSet.Int64Var(&cfg.MaxFileSizeBytes, "max-file-size", defaultMaxFileBytes, "Maximum file size in bytes.")
	flagSet.StringVar(&cfg.FailOnSeverity, "fail-on", "", "Fail if matches at or above severity.")

	if err := flagSet.Parse(args); err != nil {
		return Config{}, err
	}

	cfg.Include = include
	cfg.Exclude = exclude

	if flagSet.NArg() == 0 {
		cfg.Roots = []string{"."}
	} else {
		cfg.Roots = append([]string{}, flagSet.Args()...)
	}

	formats, err := parseFormats(*formatValue)
	if err != nil {
		return Config{}, err
	}
	cfg.Formats = formats

	if err := validateAnalyzeConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func parseFormats(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("format must not be empty")
	}

	parts := strings.Split(value, ",")
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		format := strings.TrimSpace(part)
		if format == "" {
			return nil, errors.New("format must not be empty")
		}
		if !isValidFormat(format) {
			return nil, fmt.Errorf("invalid format: %s", format)
		}
		if _, ok := seen[format]; ok {
			continue
		}
		seen[format] = struct{}{}
		result = append(result, format)
	}

	return result, nil
}

func isValidFormat(value string) bool {
	switch value {
	case "console", "json", "sarif":
		return true
	default:
		return false
	}
}

func validateAnalyzeConfig(cfg Config) error {
	if err := validateConfigPath(cfg.ConfigPath); err != nil {
		return err
	}
	if cfg.Concurrency <= 0 {
		return errors.New("concurrency must be positive")
	}
	if cfg.MaxFileSizeBytes <= 0 {
		return errors.New("max-file-size must be positive")
	}
	if cfg.FailOnSeverity != "" && !isValidSeverity(cfg.FailOnSeverity) {
		return fmt.Errorf("invalid fail-on value: %s", cfg.FailOnSeverity)
	}
	if err := validateOutputPaths(cfg); err != nil {
		return err
	}

	return nil
}

func validateConfigPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("config file not found: %s", path)
	}
	if info.IsDir() {
		return fmt.Errorf("config path is a directory: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("config file not readable: %s", path)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("config file not readable: %s", path)
	}

	return nil
}

func validateOutputPaths(cfg Config) error {
	if len(cfg.Formats) <= 1 {
		return validateOutputPathValues(cfg)
	}

	for _, format := range cfg.Formats {
		switch format {
		case "json":
			if cfg.OutJSON == "" {
				return errors.New("--out-json is required when requesting json with multiple formats")
			}
		case "sarif":
			if cfg.OutSARIF == "" {
				return errors.New("--out-sarif is required when requesting sarif with multiple formats")
			}
		}
	}

	return validateOutputPathValues(cfg)
}

func validateOutputPathValues(cfg Config) error {
	if cfg.OutJSON != "" {
		if err := validateOutputPath(cfg.OutJSON); err != nil {
			return err
		}
	}
	if cfg.OutSARIF != "" {
		if err := validateOutputPath(cfg.OutSARIF); err != nil {
			return err
		}
	}

	return nil
}

func validateOutputPath(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		return validateExistingOutputPath(info, path)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("output path not writable: %s", path)
	}

	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		parent = "."
	}
	if _, err := os.Stat(parent); err != nil {
		return fmt.Errorf("output path not writable: %s", path)
	}

	return validateOutputDirectoryWritable(parent, path)
}

func validateExistingOutputPath(info os.FileInfo, path string) error {
	if info.IsDir() {
		return fmt.Errorf("output path is a directory: %s", path)
	}
	file, err := os.OpenFile(path, os.O_WRONLY, defaultFileMode)
	if err != nil {
		return fmt.Errorf("output path not writable: %s", path)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("output path not writable: %s", path)
	}

	return nil
}

func validateOutputDirectoryWritable(parent, path string) error {
	probe, err := os.CreateTemp(parent, ".regex-checker-*")
	if err != nil {
		return fmt.Errorf("output path not writable: %s", path)
	}
	name := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(name)

		return fmt.Errorf("output path not writable: %s", path)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("output path not writable: %s", path)
	}

	return nil
}

func isValidSeverity(value string) bool {
	switch value {
	case "error", "warning", "notice", "info":
		return true
	default:
		return false
	}
}

// BuildScanRequest resolves overrides and builds a scan request.
func BuildScanRequest(cfg Config, ruleSet config.RuleSet) (scan.Request, string) {
	effective := ruleSet.ToRules()

	if len(cfg.Include) > 0 {
		effective.Include = append([]string{}, cfg.Include...)
	}
	if len(cfg.Exclude) > 0 {
		effective.Exclude = append([]string{}, cfg.Exclude...)
	}
	if cfg.FailOnSeverity != "" {
		effective.FailOn = &cfg.FailOnSeverity
	}

	effectiveRules := effective.Rules
	if len(cfg.Include) > 0 || len(cfg.Exclude) > 0 {
		effectiveRules = make([]rules.Rule, len(effective.Rules))
		for i, rule := range effective.Rules {
			copied := rule
			if len(cfg.Include) > 0 {
				copied.Paths = append([]string{}, effective.Include...)
			}
			if len(cfg.Exclude) > 0 {
				copied.Exclude = append([]string{}, effective.Exclude...)
			}
			effectiveRules[i] = copied
		}
	}

	request := scan.Request{
		Roots:            append([]string{}, cfg.Roots...),
		Rules:            effectiveRules,
		Include:          append([]string{}, effective.Include...),
		Exclude:          append([]string{}, effective.Exclude...),
		MaxFileSizeBytes: cfg.MaxFileSizeBytes,
		Concurrency:      cfg.Concurrency,
	}

	failOn := ""
	if effective.FailOn != nil {
		failOn = *effective.FailOn
	}

	return request, failOn
}

func writeError(out *bytes.Buffer, err error) {
	_, _ = fmt.Fprintf(out, "%s\n", err.Error())
}

func runAnalyze(args []string) (scan.Result, string, []string, []rules.Rule, Config, error) {
	cfg, err := ParseAnalyzeArgs(args)
	if err != nil {
		return scan.Result{}, "", nil, nil, Config{}, err
	}

	ruleSet, err := config.LoadRuleSet(cfg.ConfigPath)
	if err != nil {
		return scan.Result{}, "", nil, nil, Config{}, err
	}

	request, failOn := BuildScanRequest(cfg, ruleSet)
	result, err := scan.Run(request)
	if err != nil {
		return scan.Result{}, "", nil, nil, Config{}, err
	}

	return result, failOn, cfg.Formats, request.Rules, cfg, nil
}

func renderOutputs(formats []string, ruleset []rules.Rule, cfg Config, result scan.Result, out *bytes.Buffer) error {
	for _, format := range formats {
		switch format {
		case "console":
			if err := output.WriteConsole(result, out); err != nil {
				return err
			}
		case "json":
			if err := writeJSONOutput(cfg, result, out); err != nil {
				return err
			}
		case "sarif":
			if err := writeSARIFOutput(cfg, result, ruleset, out); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid format: %s", format)
		}
	}

	return nil
}

func writeJSONOutput(cfg Config, result scan.Result, out *bytes.Buffer) error {
	if cfg.OutJSON == "" {
		if len(cfg.Formats) != 1 {
			return errors.New("--out-json is required when requesting json with multiple formats")
		}

		return output.WriteJSON(result, out)
	}

	return writeJSONFile(cfg.OutJSON, result)
}

func writeSARIFOutput(cfg Config, result scan.Result, ruleset []rules.Rule, out *bytes.Buffer) error {
	if cfg.OutSARIF == "" {
		if len(cfg.Formats) != 1 {
			return errors.New("--out-sarif is required when requesting sarif with multiple formats")
		}

		return output.WriteSARIF(result, ruleset, out)
	}

	return writeSARIFFile(cfg.OutSARIF, result, ruleset)
}

func writeJSONFile(path string, result scan.Result) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultFileMode)
	if err != nil {
		return err
	}
	if err := output.WriteJSON(result, file); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}

func writeSARIFFile(path string, result scan.Result, ruleSet []rules.Rule) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultFileMode)
	if err != nil {
		return err
	}
	if err := output.WriteSARIF(result, ruleSet, file); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}

func hasFailingMatch(matches []scan.Match, failOn string) bool {
	threshold := severityRank(failOn)
	for _, match := range matches {
		if severityRank(match.Severity) <= threshold {
			return true
		}
	}

	return false
}

func exitCodeForFailOn(matches []scan.Match, failOn string) int {
	if hasFailingMatch(matches, failOn) {
		return exitCodeFailOn
	}

	return 0
}

const (
	severityRankError = iota
	severityRankWarning
	severityRankNotice
	severityRankInfo
	severityRankUnknown
)

func severityRank(value string) int {
	switch value {
	case "error":
		return severityRankError
	case "warning":
		return severityRankWarning
	case "notice":
		return severityRankNotice
	case "info":
		return severityRankInfo
	default:
		return severityRankUnknown
	}
}
