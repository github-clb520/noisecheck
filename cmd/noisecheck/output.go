package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"noisecheck/internal/agent"
	"noisecheck/internal/model"
	"noisecheck/internal/suggestdiff"
)

func outputText(comments []model.LlmComment) {
	if len(comments) == 0 {
		fmt.Println("✅ 审查完毕，未发现问题。")
		return
	}
	for _, c := range comments {
		renderComment(c)
	}
}

func hasSubtaskErrors(warnings []agent.AgentWarning) bool {
	for _, w := range warnings {
		if w.Type == "subtask_error" {
			return true
		}
	}
	return false
}

func outputTextWithWarnings(comments []model.LlmComment, warnings []agent.AgentWarning) {
	if len(comments) == 0 {
		if hasSubtaskErrors(warnings) {
			fmt.Println("⚠️  部分文件审查出错（详见下方警告）")
		} else {
			fmt.Println("✅ 审查完毕，未发现问题。")
		}
	} else {
		for _, c := range comments {
			renderComment(c)
		}
	}
	for _, w := range warnings {
		if w.Type == "subtask_error" {
			continue
		}
		fmt.Fprintf(os.Stderr, "[NC] WARNING [%s] %s: %s\n", w.Type, sanitizeTerminal(w.File), sanitizeTerminal(w.Message))
	}
}

// renderComment 渲染单条审查意见，含中文严重级别标签
func renderComment(comment model.LlmComment) {
	lines := buildDiffLines(comment)
	if len(lines) == 0 && comment.Content == "" {
		return
	}

	// 自动推断或使用已有严重级别
	sev := comment.Severity
	if sev == "" {
		sev = model.KeywordSeverity(comment.Content)
	}
	sevLabel, sevColor := model.SeverityChinese(sev)

	// 严重级别标签 + 文件位置
	fmt.Printf("\n%s[%s]%s \033[2m%s:%d-%d\033[0m\n",
		sevColor, sevLabel, model.ResetColor,
		sanitizeTerminal(comment.Path), comment.StartLine, comment.EndLine)

	// 分类标签（如果有）
	if comment.Category != "" {
		catLabel := categoryChinese(comment.Category)
		fmt.Printf("  \033[2m分类: %s\033[0m\n", catLabel)
	}

	// 评论内容
	if comment.Content != "" {
		for _, ln := range wrapByRunes(sanitizeTerminal(comment.Content), 100) {
			fmt.Printf("%s\n", ln)
		}
		fmt.Println()
	}

	// 建议代码 diff
	if len(lines) > 0 {
		for _, dl := range lines {
			switch dl.Type {
			case suggestdiff.DiffAdded:
				printDiffLine("+", sanitizeTerminal(dl.Content), "\033[92m", "\033[48;2;0;60;0m")
			case suggestdiff.DiffDeleted:
				printDiffLine("-", sanitizeTerminal(dl.Content), "\033[91m", "\033[48;2;70;0;0m")
			case suggestdiff.DiffContext:
				printDiffLine(" ", sanitizeTerminal(dl.Content), "\033[2m", "\033[48;2;38;38;38m")
			}
		}
	}

	fmt.Println()
}

// categoryChinese 返回分类中文名
func categoryChinese(cat string) string {
	switch strings.ToLower(cat) {
	case "security":
		return "安全"
	case "performance":
		return "性能"
	case "correctness":
		return "正确性"
	case "maintainability":
		return "可维护性"
	default:
		return cat
	}
}

// printDiffLine renders a single diff line with colored prefix and background on content.
func printDiffLine(prefix, content, fgColor, bgColor string) {
	fmt.Printf("%s%s%s %s%s\033[0m\n", fgColor+bgColor, prefix, "\033[0m"+bgColor, content, "\033[0m")
}

// wrapByRunes splits text into lines that fit within maxWidth **rune** columns.
// Respects existing newlines and wraps at word boundaries.
func wrapByRunes(text string, maxW int) []string {
	if text == "" {
		return nil
	}
	var result []string
	for _, para := range strings.Split(text, "\n") {
		result = append(result, wrapSingleRuneLine(para, maxW)...)
	}
	return result
}

// wrapSingleRuneLine breaks one paragraph into rune-width-constrained lines.
func wrapSingleRuneLine(line string, maxW int) []string {
	runes := []rune(line)
	if visibleRunesLen(runes) <= maxW {
		return []string{line}
	}
	var result []string
	for len(runes) > 0 {
		cut := runeWrapCut(runes, maxW)
		result = append(result, string(runes[:cut]))
		runes = runes[cut:]
		for len(runes) > 0 && runes[0] == ' ' {
			runes = runes[1:]
		}
	}
	return result
}

// runeWrapCut returns a rune index suitable for breaking the line at ~maxW display width.
func runeWrapCut(runes []rune, maxW int) int {
	if visibleRunesLen(runes) <= maxW {
		return len(runes)
	}
	best := maxW
	if best >= len(runes) {
		return len(runes)
	}
	for i := best; i > 0; i-- {
		if runes[i] == ' ' || runes[i] == '\t' {
			return i
		}
	}
	return best
}

func visibleRunesLen(runes []rune) int {
	n := 0
	for _, r := range runes {
		if r >= 32 && r != 127 {
			n++
		}
	}
	return n
}

func sanitizeTerminal(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' || r == '\n' || !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func splitToLines(s string) []string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func buildDiffLines(comment model.LlmComment) []suggestdiff.DiffLine {
	if comment.SuggestionCode == "" || comment.ExistingCode == "" {
		return nil
	}
	oldLines := splitToLines(comment.ExistingCode)
	newLines := splitToLines(comment.SuggestionCode)
	return suggestdiff.ComputeLineDiff(oldLines, newLines)
}

// --- JSON output (unchanged) ---

type jsonSummary struct {
	FilesReviewed    int64  `json:"files_reviewed"`
	Comments         int64  `json:"comments"`
	TotalTokens      int64  `json:"total_tokens"`
	InputTokens      int64  `json:"input_tokens"`
	OutputTokens     int64  `json:"output_tokens"`
	CacheReadTokens  int64  `json:"cache_read_tokens,omitempty"`
	CacheWriteTokens int64  `json:"cache_write_tokens,omitempty"`
	Elapsed          string `json:"elapsed"`
}

type jsonOutput struct {
	Status   string               `json:"status"`
	Message  string               `json:"message,omitempty"`
	Summary  *jsonSummary         `json:"summary,omitempty"`
	Comments []model.LlmComment   `json:"comments"`
	Warnings []agent.AgentWarning `json:"warnings,omitempty"`
}

func outputJSON(comments []model.LlmComment) error {
	out := jsonOutput{
		Status:   "success",
		Comments: comments,
	}
	if len(comments) == 0 {
		out.Message = "No comments generated. Looks good to me."
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func outputJSONWithWarnings(comments []model.LlmComment, warnings []agent.AgentWarning,
	filesReviewed, inputTokens, outputTokens, totalTokens, cacheReadTokens, cacheWriteTokens int64, duration time.Duration) error {
	out := jsonOutput{
		Status:   "success",
		Comments: comments,
		Summary: &jsonSummary{
			FilesReviewed:    filesReviewed,
			Comments:         int64(len(comments)),
			TotalTokens:      totalTokens,
			InputTokens:      inputTokens,
			OutputTokens:     outputTokens,
			CacheReadTokens:  cacheReadTokens,
			CacheWriteTokens: cacheWriteTokens,
			Elapsed:          duration.Round(time.Second).String(),
		},
	}
	if len(comments) == 0 {
		if hasSubtaskErrors(warnings) {
			out.Message = "Some files could not be reviewed due to errors."
		} else {
			out.Message = "No comments generated. Looks good to me."
		}
	}
	if len(warnings) > 0 {
		out.Warnings = warnings
		if hasSubtaskErrors(warnings) {
			out.Status = "completed_with_errors"
		} else {
			out.Status = "completed_with_warnings"
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func outputJSONNoFiles() error {
	out := jsonOutput{
		Status:   "skipped",
		Message:  "No supported files changed.",
		Comments: []model.LlmComment{},
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// --- Markdown output (适用于 CI 日志) ---

func outputMarkdown(comments []model.LlmComment) {
	if len(comments) == 0 {
		fmt.Println("✅ No issues found.")
		return
	}
	fmt.Printf("## NoiseCheck 审查报告\n\n")
	fmt.Printf("共发现 **%d** 个问题\n\n", len(comments))

	// 按文件分组
	byFile := make(map[string][]model.LlmComment)
	for _, c := range comments {
		path := c.Path
		if path == "" {
			path = "(unknown)"
		}
		byFile[path] = append(byFile[path], c)
	}

	for path, pathComments := range byFile {
		fmt.Printf("### %s\n\n", path)
		for i, c := range pathComments {
			sev := c.Severity
			if sev == "" {
				sev = model.KeywordSeverity(c.Content)
			}
			sevLabel, _ := model.SeverityChinese(sev)
			lines := ""
			if c.StartLine > 0 {
				lines = fmt.Sprintf(" L%d-%d", c.StartLine, c.EndLine)
			}
			fmt.Printf("%d. **`[%s]`** `%s%s`\n", i+1, sevLabel, path, lines)
			fmt.Printf("   %s\n", c.Content)
			if c.SuggestionCode != "" {
				fmt.Printf("   ```suggestion\n   %s\n   ```\n", c.SuggestionCode)
			}
			fmt.Println()
		}
	}
}

func outputMarkdownWithWarnings(comments []model.LlmComment, warnings []agent.AgentWarning,
	filesReviewed, inputTokens, outputTokens, totalTokens, cacheReadTokens, cacheWriteTokens int64, duration time.Duration) {
	outputMarkdown(comments)
	if len(warnings) > 0 {
		fmt.Printf("### 警告\n\n")
		for _, w := range warnings {
			if w.Type == "subtask_error" {
				continue
			}
			fmt.Printf("- `[%s]` %s: %s\n", w.Type, w.File, w.Message)
		}
		fmt.Println()
	}
	fmt.Printf("---\n*审查文件数: %d | 总计消耗: ~%d tokens | 耗时: %s*\n",
		filesReviewed, totalTokens, duration.Round(time.Second))
}

// --- Preview output ---

func outputPreviewText(p *agent.DiffPreview) {
	if p.TotalFiles == 0 {
		fmt.Println("No files changed.")
		return
	}

	maxPathLen := 0
	for _, e := range p.Entries {
		if n := len(sanitizeTerminal(e.Path)); n > maxPathLen {
			maxPathLen = n
		}
	}
	if maxPathLen < 20 {
		maxPathLen = 20
	}
	pathFmt := fmt.Sprintf("%%-%ds", maxPathLen)

	fmt.Printf("\nPreview: %d file(s) changed  |  \033[32m+%d\033[0m  \033[31m-%d\033[0m\n",
		p.TotalFiles, p.TotalInsertions, p.TotalDeletions)

	if p.ReviewableCount > 0 {
		fmt.Printf("\n\033[1mWill review (%d):\033[0m\n", p.ReviewableCount)
		for _, e := range p.Entries {
			if !e.WillReview {
				continue
			}
			fmt.Printf("  %s  "+pathFmt+" \033[32m+%-4d\033[0m \033[31m-%-4d\033[0m\n",
				statusBadge(e.Status), sanitizeTerminal(e.Path), e.Insertions, e.Deletions)
		}
	}

	if p.ExcludedCount > 0 {
		fmt.Printf("\n\033[1mExcluded from review (%d):\033[0m\n", p.ExcludedCount)
		for _, e := range p.Entries {
			if e.WillReview {
				continue
			}
			fmt.Printf("  %s  "+pathFmt+" \033[2m(%s)\033[0m\n",
				statusBadge(e.Status), sanitizeTerminal(e.Path), sanitizeTerminal(string(e.ExcludeReason)))
		}
	}

	fmt.Println()
}

func statusBadge(status string) string {
	switch status {
	case "added":
		return "\033[32m[A]\033[0m"
	case "modified":
		return "\033[33m[M]\033[0m"
	case "deleted":
		return "\033[31m[D]\033[0m"
	case "renamed":
		return "\033[36m[R]\033[0m"
	case "binary":
		return "\033[35m[B]\033[0m"
	default:
		return "[?]"
	}
}
