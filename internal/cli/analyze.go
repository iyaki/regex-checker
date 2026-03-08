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

	"github.com/iyaki/reglint/internal/config"
	"github.com/iyaki/reglint/internal/output"
	"github.com/iyaki/reglint/internal/rules"
	"github.com/iyaki/reglint/internal/scan"
)

const (
	defaultConfigPath   = "reglint-rules.yaml"
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
	ConcurrencySet   bool
	MaxFileSizeBytes int64
	FailOnSeverity   string
	NoIgnoreFiles    bool
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
	result, failOn, formats, ruleset, cfg, consoleColors, err := runAnalyze(args)
	if err != nil {
		writeError(out, err)

		return exitCodeError
	}

	if err := renderOutputs(formats, ruleset, cfg, consoleColors, result, out); err != nil {
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
	flagSet.StringVar(&cfg.ConfigPath, "c", defaultConfigPath, "Path to YAML rules config file.")
	formatValue := flagSet.String("format", defaultFormat, "Comma-separated list of formats.")
	formatShort := flagSet.String("f", defaultFormat, "Comma-separated list of formats.")
	flagSet.StringVar(&cfg.OutJSON, "out-json", "", "Output path for JSON results.")
	flagSet.StringVar(&cfg.OutSARIF, "out-sarif", "", "Output path for SARIF results.")
	var include stringSlice
	var exclude stringSlice
	flagSet.Var(&include, "include", "Repeatable include glob.")
	flagSet.Var(&exclude, "exclude", "Repeatable exclude glob.")
	flagSet.IntVar(&cfg.Concurrency, "concurrency", runtime.GOMAXPROCS(0), "Worker count.")
	flagSet.Int64Var(&cfg.MaxFileSizeBytes, "max-file-size", defaultMaxFileBytes, "Maximum file size in bytes.")
	flagSet.StringVar(&cfg.FailOnSeverity, "fail-on", "", "Fail if matches at or above severity.")
	flagSet.BoolVar(&cfg.NoIgnoreFiles, "no-ignore-files", false, "Disable ignore file loading and matching.")

	if err := flagSet.Parse(args); err != nil {
		return Config{}, err
	}

	cfg.Include = include
	cfg.Exclude = exclude
	if wasFlagProvided(flagSet, "concurrency") {
		cfg.ConcurrencySet = true
	}

	if flagSet.NArg() == 0 {
		cfg.Roots = []string{"."}
	} else {
		cfg.Roots = append([]string{}, flagSet.Args()...)
	}

	formatInput := *formatValue
	if wasFlagProvided(flagSet, "f") && !wasFlagProvided(flagSet, "format") {
		formatInput = *formatShort
	}
	formats, err := parseFormats(formatInput)
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
		if _, ok := seen[format]; ok {
			continue
		}
		seen[format] = struct{}{}
		result = append(result, format)
	}

	registry, err := output.NewRegistry(
		output.ConsoleFormatter{},
		output.JSONFormatter{},
		output.SARIFFormatter{},
	)
	if err != nil {
		return nil, err
	}
	if _, err := registry.Resolve(result); err != nil {
		return nil, err
	}

	return result, nil
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
	file, err := os.Open(path) //#nosec G304 -- path comes from user input and is validated here
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
	file, err := os.OpenFile(path, os.O_WRONLY, defaultFileMode) //#nosec G304 -- validated output path
	if err != nil {
		return fmt.Errorf("output path not writable: %s", path)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("output path not writable: %s", path)
	}

	return nil
}

func validateOutputDirectoryWritable(parent, path string) error {
	probe, err := os.CreateTemp(parent, ".reglint-*")
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
func BuildScanRequest(cfg Config, ruleSet config.RuleSet) (scan.Request, string, output.ConsoleColorSettings) {
	effective := ruleSet.ToRules()

	applyRuleSetOverrides(cfg, &effective)

	ignoreSettings := resolveIgnoreSettings(effective)
	consoleColorSettings := resolveConsoleColorSettings(effective)

	request := scan.Request{
		Roots:            append([]string{}, cfg.Roots...),
		Rules:            buildEffectiveRules(cfg, effective),
		Include:          append([]string{}, effective.Include...),
		Exclude:          append([]string{}, effective.Exclude...),
		Ignore:           ignoreSettings,
		MaxFileSizeBytes: cfg.MaxFileSizeBytes,
		Concurrency:      resolveConcurrency(cfg, effective.Concurrency),
	}

	return request, resolveFailOn(effective.FailOn), consoleColorSettings
}

func writeError(out *bytes.Buffer, err error) {
	_, _ = fmt.Fprintf(out, "%s\n", err.Error())
}

func runAnalyze(
	args []string,
) (scan.Result, string, []string, []rules.Rule, Config, output.ConsoleColorSettings, error) {
	cfg, err := ParseAnalyzeArgs(args)
	if err != nil {
		return scan.Result{}, "", nil, nil, Config{}, output.ConsoleColorSettings{}, err
	}

	ruleSet, err := config.LoadRuleSet(cfg.ConfigPath)
	if err != nil {
		return scan.Result{}, "", nil, nil, Config{}, output.ConsoleColorSettings{}, err
	}

	request, failOn, consoleColorSettings := BuildScanRequest(cfg, ruleSet)
	result, err := scan.Run(request)
	if err != nil {
		return scan.Result{}, "", nil, nil, Config{}, output.ConsoleColorSettings{}, err
	}

	return result, failOn, cfg.Formats, request.Rules, cfg, consoleColorSettings, nil
}

func renderOutputs(
	formats []string,
	ruleset []rules.Rule,
	cfg Config,
	consoleColors output.ConsoleColorSettings,
	result scan.Result,
	out *bytes.Buffer,
) error {
	registry, err := outputRegistry(ruleset, consoleColors)
	if err != nil {
		return err
	}

	for _, format := range formats {
		formatter, err := registry.ResolveName(format)
		if err != nil {
			return err
		}
		if err := renderFormat(formatter, cfg, ruleset, result, out); err != nil {
			return err
		}
	}

	return nil
}

var outputRegistry = defaultOutputRegistry

func defaultOutputRegistry(ruleset []rules.Rule, consoleColors output.ConsoleColorSettings) (*output.Registry, error) {
	return output.NewRegistry(
		output.ConsoleFormatter{ColorSettings: consoleColors},
		output.JSONFormatter{},
		output.SARIFFormatter{Rules: ruleset},
	)
}

func renderFormat(
	formatter output.Formatter,
	cfg Config,
	ruleset []rules.Rule,
	result scan.Result,
	out *bytes.Buffer,
) error {
	switch formatter.Name() {
	case "console":
		return formatter.Write(result, out)
	case "json":
		return writeJSONOutput(cfg, result, out)
	case "sarif":
		return writeSARIFOutput(cfg, result, ruleset, out)
	default:
		return fmt.Errorf("invalid format: %s", formatter.Name())
	}
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
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultFileMode) //#nosec G304
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
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultFileMode) //#nosec G304
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

func applyRuleSetOverrides(cfg Config, effective *rules.RuleSet) {
	if len(cfg.Include) > 0 {
		effective.Include = append([]string{}, cfg.Include...)
	}
	if len(cfg.Exclude) > 0 {
		effective.Exclude = append([]string{}, cfg.Exclude...)
	}
	if cfg.FailOnSeverity != "" {
		effective.FailOn = &cfg.FailOnSeverity
	}
	if cfg.NoIgnoreFiles {
		ignoreFilesDisabled := false
		effective.IgnoreFilesEnabled = &ignoreFilesDisabled
	}
}

func buildEffectiveRules(cfg Config, effective rules.RuleSet) []rules.Rule {
	if len(cfg.Include) == 0 && len(cfg.Exclude) == 0 {
		return effective.Rules
	}

	effectiveRules := make([]rules.Rule, len(effective.Rules))
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

	return effectiveRules
}

func resolveConcurrency(cfg Config, rulesetConcurrency *int) int {
	if !cfg.ConcurrencySet && rulesetConcurrency != nil {
		return *rulesetConcurrency
	}

	return cfg.Concurrency
}

func resolveFailOn(failOn *string) string {
	if failOn == nil {
		return ""
	}

	return *failOn
}

func resolveConsoleColorSettings(effective rules.RuleSet) output.ConsoleColorSettings {
	settings := output.ConsoleColorSettings{
		Enabled: true,
		Source:  output.ConsoleColorSourceDefault,
	}
	if effective.ConsoleColorsEnabled != nil {
		settings.Enabled = *effective.ConsoleColorsEnabled
		settings.Source = output.ConsoleColorSourceConfig
	}

	if envValue, ok := os.LookupEnv("NO_COLOR"); ok && envValue != "" {
		settings.Enabled = false
		settings.Source = output.ConsoleColorSourceEnv
	}

	return settings
}

func resolveIgnoreSettings(effective rules.RuleSet) scan.IgnoreSettings {
	settings := scan.IgnoreSettings{
		Enabled: true,
		Files:   []string{".ignore", ".reglintignore"},
	}
	if effective.IgnoreFilesEnabled != nil {
		settings.Enabled = *effective.IgnoreFilesEnabled
	}
	if len(effective.IgnoreFiles) > 0 {
		settings.Files = append([]string{}, effective.IgnoreFiles...)
	}

	return settings
}

func wasFlagProvided(flagSet *flag.FlagSet, name string) bool {
	found := false
	flagSet.Visit(func(flagItem *flag.Flag) {
		if flagItem.Name == name {
			found = true
		}
	})

	return found
}
