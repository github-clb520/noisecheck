// NoiseCheck is an AI-powered code review CLI tool.
// It reads git diffs, sends them to a configurable LLM service, and generates review comments.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"noisecheck/internal/llm"
	"noisecheck/internal/telemetry"
)

func main() {
	llm.AppVersion = Version
	llm.InitEmbeddedLoader()

	ctx := context.Background()
	if telemetry.Init(ctx) {
		defer telemetry.ShutdownWithTimeout(ctx, 5*time.Second)
	}

	if err := dispatch(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// dispatch routes top-level subcommands or global flags.
func dispatch() error {
	args := os.Args[1:]

	// No args → default to review with empty args (will trigger usage/help)
	if len(args) == 0 {
		printTopLevelUsage()
		return nil
	}

	switch args[0] {
	case "--version", "-V":
		printVersion()
		return nil
	case "version":
		printVersion()
		return nil
	case "review", "r":
		return runReview(args[1:])
	case "init":
		return runInit(args[1:])
	case "config":
		return runConfig(args[1:])
	case "llm":
		return runLLM(args[1:])
	case "rules":
		return runRules(args[1:])
	case "viewer":
		return runViewer(args[1:])
	case "-h", "--help":
		printTopLevelUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s\nRun 'nc' for usage", args[0])
	}
}

func printTopLevelUsage() {
	fmt.Println(`NoiseCheck - AI-Powered Code Review CLI

Usage:
  nc [command]

Commands:
  init         Initialize configuration (interactive wizard)
  review, r    Start a code review
  rules        Inspect and debug review rules
  config       Manage configuration settings
  llm          LLM utility commands
  viewer       Start the WebUI session viewer
  version      Show version information

Examples:
  nc init                                Interactive setup wizard
  nc review --from master --to dev       Review diff range
  nc review --commit abc123              Review a single commit
  nc review --format markdown            CI-friendly markdown output
  nc review --report report.html         Generate HTML report
  nc config provider                     Interactive provider setup
  nc config model                        Interactive model selection
  nc config set llm.model opus-4-6       Set a config value
  nc llm test                            Test LLM connectivity
  nc llm providers                       List built-in providers
  nc version                             Show version info

Use "nc init" for interactive setup.
Use "nc review -h" for more information about review.
Use "nc rules -h" for more information about rules.
Use "nc config" for more information about config.
Use "nc llm" for more information about LLM utilities.

GitHub: https://github.com/github-clb520/noisecheck`)
}
