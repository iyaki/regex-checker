package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	defaultInitPath = "regex-rules.yaml"
)

// InitConfig holds parsed init command inputs.
type InitConfig struct {
	OutputPath string
	Force      bool
}

// HandleInit executes the init command.
func HandleInit(args []string, out *bytes.Buffer) int {
	cfg, err := ParseInitArgs(args)
	if err != nil {
		writeError(out, err)

		return exitCodeError
	}

	if err := writeDefaultConfig(cfg); err != nil {
		writeError(out, err)

		return exitCodeError
	}

	_, _ = fmt.Fprintf(out, "Wrote default config to %s\n", cfg.OutputPath)

	return 0
}

// ParseInitArgs parses init command arguments into a config.
func ParseInitArgs(args []string) (InitConfig, error) {
	var cfg InitConfig

	flagSet := flag.NewFlagSet("init", flag.ContinueOnError)
	flagSet.SetOutput(&strings.Builder{})

	flagSet.StringVar(&cfg.OutputPath, "out", defaultInitPath, "Output path for the config file.")
	flagSet.BoolVar(&cfg.Force, "force", false, "Overwrite existing config file.")

	if err := flagSet.Parse(args); err != nil {
		return InitConfig{}, err
	}

	if strings.TrimSpace(cfg.OutputPath) == "" {
		return InitConfig{}, errors.New("output path must not be empty")
	}

	return cfg, nil
}

func writeDefaultConfig(cfg InitConfig) error {
	if !cfg.Force {
		if _, err := os.Stat(cfg.OutputPath); err == nil {
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", cfg.OutputPath)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	return os.WriteFile(cfg.OutputPath, []byte(defaultConfigTemplate()), defaultFileMode)
}

func defaultConfigTemplate() string {
	return "" +
		"include:\n" +
		"  - \"**/*\"\n" +
		"exclude:\n" +
		"  - \"**/.git/**\"\n" +
		"  - \"**/node_modules/**\"\n" +
		"  - \"**/vendor/**\"\n" +
		"failOn: \"error\"\n" +
		"rules:\n" +
		"  - message: \"Avoid hardcoded token: $1\"\n" +
		"    regex: \"token\\s*[:=]\\s*([A-Za-z0-9_-]+)\"\n" +
		"    severity: \"error\"\n" +
		"    paths:\n" +
		"      - \"src/**\"\n"
}
