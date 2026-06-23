package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// --- custom flag set that supports short flags (-c, -f etc.) ---

type ncFlagSet struct {
	fs       *flag.FlagSet
	shortMap map[string]string // maps short key "c" -> full name "commit"
	showHelp bool
}

func newNcFlagSet(name string) *ncFlagSet {
	return &ncFlagSet{
		fs:       flag.NewFlagSet(name, flag.ContinueOnError),
		shortMap: make(map[string]string),
	}
}

// StringVarP registers --name with optional short form -s.
func (a *ncFlagSet) StringVarP(p *string, name, shorthand string, value, usage string) {
	suffix := ""
	if shorthand != "" {
		a.shortMap[shorthand] = name
		suffix = fmt.Sprintf(" (shorthand: -%s)", shorthand)
	}
	a.fs.StringVar(p, name, value, usage+suffix)
}

// BoolVarP registers --name with optional short form -s.
func (a *ncFlagSet) BoolVarP(p *bool, name, shorthand string, value bool, usage string) {
	suffix := ""
	if shorthand != "" {
		a.shortMap[shorthand] = name
		suffix = fmt.Sprintf(" (shorthand: -%s)", shorthand)
	}
	a.fs.BoolVar(p, name, value, usage+suffix)
}

func (a *ncFlagSet) StringVar(p *string, name string, value string, usage string) {
	a.fs.StringVar(p, name, value, usage)
}

func (a *ncFlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	a.fs.BoolVar(p, name, value, usage)
}

func (a *ncFlagSet) IntVar(p *int, name string, value int, usage string) {
	a.fs.IntVar(p, name, value, usage)
}

func (a *ncFlagSet) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	a.fs.DurationVar(p, name, value, usage)
}

func (a *ncFlagSet) PrintDefaults() {
	a.fs.PrintDefaults()
}

func (a *ncFlagSet) Parse(arguments []string) error {
	expanded := expandShortFlags(arguments, a.shortMap)

	for _, arg := range expanded {
		if arg == "-h" || arg == "--help" {
			a.showHelp = true
			return nil
		}
	}

	return a.fs.Parse(expanded)
}

// expandShortFlags replaces standalone -X args with their long equivalents.
// Only triggers when the arg is exactly -N (single char after dash).
func expandShortFlags(args []string, shortMap map[string]string) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		if len(arg) == 2 && arg[0] == '-' && arg[1] != '-' {
			key := string(arg[1])
			if full, ok := shortMap[key]; ok {
				out = append(out, "--"+full)
				continue
			}
		}
		out = append(out, arg)
	}
	return out
}

// --- review subcommand options ---

type reviewOptions struct {
	toolConfigPath string
	rulePath       string
	repoDir        string
	from           string
	to             string
	commit         string
	outputFormat   string
	audience       string // --audience: "human" (default) or "agent"
	background     string // --background: optional requirement context
	model          string // --model: override resolved LLM model for this review
	concurrency    int
	perFileTimeout int
	maxTools       int
	maxGitProcs    int
	preview        bool
	reportPath     string
	showHelp       bool
}

func parseReviewFlags(args []string) (reviewOptions, error) {
	a := newNcFlagSet("nc review")

	opts := reviewOptions{}

	a.StringVar(&opts.toolConfigPath, "tools", "", "path to JSON tools config file (default: embedded)")
	a.StringVar(&opts.rulePath, "rule", "", "path to JSON file with system review rules")
	a.StringVar(&opts.repoDir, "repo", "", "root directory of the git repository (default: current dir)")
	a.StringVar(&opts.from, "from", "", "source ref to start diff from (e.g., 'main')")
	a.StringVar(&opts.to, "to", "", "target ref to end diff at (e.g., 'feature-branch')")
	a.StringVarP(&opts.commit, "commit", "c", "", "single commit hash or tag to review (vs its parent)")
	a.StringVarP(&opts.outputFormat, "format", "f", "text", "output format: text, json or markdown")
	a.IntVar(&opts.concurrency, "concurrency", 8, "max concurrent file reviews")
	a.IntVar(&opts.perFileTimeout, "timeout", 10, "concurrent task timeout in minutes")
	a.StringVar(&opts.audience, "audience", "human", "output audience: human (show progress) or agent (summary only)")
	a.StringVarP(&opts.background, "background", "b", "", "optional requirement/business context for the review")
	a.StringVar(&opts.model, "model", "", "override LLM model for this review (e.g., claude-opus-4-6)")
	a.IntVar(&opts.maxTools, "max-tools", 0, "max tool call rounds per file (0 = template default; min 10)")
	a.IntVar(&opts.maxGitProcs, "max-git-procs", 16, "max concurrent git subprocesses")
	a.BoolVarP(&opts.preview, "preview", "p", false, "preview which files will be reviewed without running the LLM")
	a.StringVar(&opts.reportPath, "report", "", "generate HTML report at this path (e.g., report.html)")

	if err := a.Parse(args); err != nil {
		return opts, fmt.Errorf("parse flags: %w", err)
	}

	opts.showHelp = a.showHelp
	if opts.showHelp {
		return opts, nil
	}

	modeCount := 0
	if opts.from != "" || opts.to != "" {
		modeCount++
	}
	if opts.commit != "" {
		modeCount++
	}
	// modeCount == 0 → workspace mode (no error, allowed)
	if modeCount > 1 {
		return opts, fmt.Errorf("only one review mode allowed (--from/--to or --commit)")
	}
	if opts.from != "" && opts.to == "" {
		return opts, fmt.Errorf("--to is required when --from is specified")
	}
	if opts.to != "" && opts.from == "" {
		return opts, fmt.Errorf("--from is required when --to is specified")
	}

	switch opts.audience {
	case "human", "agent":
	default:
		return opts, fmt.Errorf("invalid --audience value %q: must be 'human' or 'agent'", opts.audience)
	}

	switch opts.outputFormat {
	case "text", "json", "markdown":
	default:
		return opts, fmt.Errorf("invalid --format value %q: must be 'text', 'json' or 'markdown'", opts.outputFormat)
	}

	const minMaxTools = 10
	if opts.maxTools < 0 {
		return opts, fmt.Errorf("--max-tools must be a non-negative integer (0 means use template default)")
	}
	if opts.maxTools > 0 && opts.maxTools < minMaxTools {
		fmt.Fprintf(os.Stderr, "[NC] --max-tools %d is below minimum %d, using %d\n", opts.maxTools, minMaxTools, minMaxTools)
		opts.maxTools = minMaxTools
	}

	if opts.maxGitProcs < 0 {
		return opts, fmt.Errorf("--max-git-procs must be a non-negative integer (0 means use default 16)")
	}

	return opts, nil
}

func printReviewUsage() {
	fmt.Println(`NoiseCheck - AI-Powered Code Review CLI

Usage:
  nc review [flags]
  nc r [flags]                (alias)

Examples:
  # Review staged + unstaged + untracked changes in current workspace
  nc review

  # Review a branch against its base (merge-base mode)
  nc review --from master --to dev-ref

  # Review a specific commit
  nc review --commit abc123
  nc review -c abc123

  # Output JSON format
  nc review --format json
  nc review -f json

  # Output Markdown format (CI-friendly)
  nc review --format markdown

  # Generate HTML report
  nc review --report report.html

  # Agent mode (summary only, no progress lines)
  nc review --audience agent

  # Preview which files will be reviewed
  nc review --preview
  nc review -c abc123 -p

Flags:
  --audience string       output audience: human (show progress) or agent (summary only) (default "human")
  -b, --background string optional requirement/business context for the review
  -c, --commit string     single commit hash or tag to review (vs its parent)
  -f, --format string     output format: text, json or markdown (default "text")
  --concurrency int       max concurrent file reviews (default 8)
  --max-git-procs int     max concurrent git subprocesses (default 16)
  --from string           source ref to start diff from (e.g., 'main')
  --max-tools int         max tool call rounds per file (0 = template default; min 10)
  --model string          override LLM model for this review (e.g., claude-opus-4-6)
  -p, --preview           preview which files will be reviewed without running the LLM
  --report string         generate HTML report at this path (e.g., report.html)
  --repo string           root directory of the git repository (default: current dir)
  --rule string           path to JSON file with system review rules
  --timeout int           concurrent task timeout in minutes (default 10)
  --to string             target ref to end diff at (e.g., 'feature-branch')
  --tools string          path to JSON tools config file (default: embedded)`)
}

// --- config subcommand ---

type configAction struct {
	subCmd string // "set"
	key    string
	value  string
}

func parseConfigArgs(args []string) (configAction, error) {
	if len(args) == 0 {
		return configAction{}, fmt.Errorf("usage: ocr config set <key> <value>\ne.g., ocr config set llm.model claude-opus-4-6")
	}

	subCmd := args[0]
	switch subCmd {
	case "set":
		if len(args) < 3 {
			return configAction{}, fmt.Errorf("usage: nc config set <key> <value>\ne.g., nc config set llm.model claude-opus-4-6")
		}
		return configAction{
			subCmd: "set",
			key:    args[1],
			value:  args[2],
		}, nil
	default:
		return configAction{}, fmt.Errorf("unknown config sub-command: %s\nAvailable: set, provider, model", subCmd)
	}
}

func printConfigUsage() {
	fmt.Println(`Configuration management.

Usage:
  nc config set <key> <value>
  nc config provider              Interactive provider setup
  nc config model                 Interactive model selection

Examples:
  # Provider setup (interactive)
  nc config provider
  nc config model

  # Provider setup (non-interactive)
  nc config set provider anthropic
  nc config set model claude-opus-4-6
  # Set API key via environment variable (recommended) or config:
  # export ANTHROPIC_API_KEY=sk-ant-xxx
  nc config set providers.anthropic.api_key "$ANTHROPIC_API_KEY"

  # Custom provider
  nc config set provider my-gateway
  nc config set custom_providers.my-gateway.url https://gateway.internal.com/v1
  nc config set custom_providers.my-gateway.protocol openai
  nc config set custom_providers.my-gateway.model llama-3-70b
  nc config set custom_providers.my-gateway.models '["llama-3-70b","llama-3-8b"]'
  nc config set custom_providers.my-gateway.api_key "$MY_API_KEY"

  # Legacy endpoint configuration
  nc config set llm.url https://xx/v1/openai/chat/completions
  nc config set llm.auth_token xxxxxxxxxx
  nc config set llm.auth_header x-api-key
  nc config set llm.model claude-opus-4-6
  nc config set llm.extra_body '{"thinking":{"type":"disabled"}}'
  nc config set language English
  nc config set telemetry.enabled true

Supported keys: provider, model, providers.<name>.<field>, custom_providers.<name>.<field>, llm.url, llm.auth_token, llm.auth_header, llm.model, llm.use_anthropic, llm.extra_body, language, telemetry.enabled, telemetry.exporter, telemetry.otlp_endpoint, telemetry.content_logging
Provider fields: api_key, url, protocol, model, models, auth_header, extra_body`)
}
