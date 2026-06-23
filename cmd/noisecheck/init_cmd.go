package main

import (
	"fmt"
	"os"
)

func runInit(args []string) error {
	// Check if already configured
	configPath, err := defaultConfigPath()
	if err != nil {
		return err
	}

	cfg, err := loadOrCreateConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// If config exists and has LLM set, ask if user wants to reconfigure
	if cfg.Provider != "" || cfg.Llm.URL != "" {
		fmt.Println("NoiseCheck 已经配置过 LLM 提供商。")
		fmt.Println("如需重新配置，请运行: nc config provider")
		fmt.Println("跳转到其他设置...")
	}

	// Check if we can run the TUI
	if !isInteractive() {
		fmt.Println("非交互式环境，跳过初始化向导。")
		fmt.Println("使用 'nc config provider' 手动配置 LLM。")
		return nil
	}

	// Run the TUI init wizard
	return runInitWizard(configPath, cfg)
}

func runInitWizard(configPath string, cfg *Config) error {
	fmt.Print(`
╔══════════════════════════════════════════════╗
║         🎯 NoiseCheck 初始化向导              ║
║     零噪音 AI 代码审查 CLI                    ║
╚══════════════════════════════════════════════╝
`)
	fmt.Println("本向导将帮助你在 2 分钟内完成配置。")

	// Step 1: Language selection
	lang := askChoice("审查语言", []string{"中文（推荐）", "English"}, "中文（推荐）")
	switch lang {
	case "中文（推荐）":
		cfg.Language = "Chinese"
	case "English":
		cfg.Language = "English"
	}

	// Step 2: Review level
	level := askChoice("审查严格程度", []string{
		"标准 - 平衡质量和性能（推荐）",
		"严格 - 最全面审查（最多 LLM 调用）",
		"轻量 - 快速扫描（适合 CI）",
	}, "标准 - 平衡质量和性能（推荐）")

	_ = level // Will be used when we add level config

	// Step 3: API Key setup
	fmt.Println("\n📡 LLM 提供商配置")
	fmt.Println("支持: Anthropic Claude / OpenAI / 兼容 API")

	setupMethod := askChoice("如何配置 LLM？", []string{
		"Anthropic Claude（推荐）",
		"OpenAI / 兼容 API",
		"跳过（稍后配置）",
	}, "Anthropic Claude（推荐）")

	switch setupMethod {
	case "Anthropic Claude（推荐）":
		return runConfigProvider()
	case "OpenAI / 兼容 API":
		return runConfigProvider() // TUI now supports custom/manual
	case "跳过（稍后配置）":
		fmt.Println("\n⏭️  跳过 LLM 配置。运行 'nc config provider' 稍后配置。")
		return nil
	}

	return nil
}

// isInteractive returns true if we're in a terminal.
func isInteractive() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// askChoice presents options and returns the selection (full text).
func askChoice(prompt string, options []string, defaultOption string) string {
	if !isInteractive() {
		return defaultOption
	}

	fmt.Printf("\n%s:\n", prompt)
	for i, opt := range options {
		mark := " "
		if opt == defaultOption {
			mark = "▶"
		}
		fmt.Printf("  %s %d) %s\n", mark, i+1, opt)
	}

	var choice int
	fmt.Printf("\n请选择 [1-%d] (默认 %d): ", len(options), indexOf(options, defaultOption)+1)
	_, err := fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(options) {
		return defaultOption
	}
	return options[choice-1]
}

func indexOf(items []string, item string) int {
	for i, s := range items {
		if s == item {
			return i
		}
	}
	return -1
}
