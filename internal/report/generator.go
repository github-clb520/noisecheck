// Package report generates HTML review reports from LlmComment results.
package report

import (
	"embed"
	"html"
	"html/template"
	"os"
	"sort"
	"strings"
	"time"

	"noisecheck/internal/model"
)

//go:embed template.html
var reportTemplate embed.FS

// severityLabel returns Chinese severity labels
var severityLabel = map[string]string{
	model.SeverityCritical: "严重",
	model.SeverityHigh:     "高危",
	model.SeverityMedium:   "中危",
	model.SeverityLow:      "低危",
}

// categoryLabel returns Chinese category labels
var categoryLabel = map[string]string{
	"security":       "安全",
	"performance":    "性能",
	"correctness":    "正确性",
	"maintainability": "可维护性",
}

// ViewComment is the template-friendly view model for a single comment.
type ViewComment struct {
	Severity       string
	SevLabel       string
	Category       string
	Content        string
	Lines          string
	SuggestionHTML template.HTML
	ExistingHTML   template.HTML
}

// FileGroup groups comments by file path.
type FileGroup struct {
	ID       string
	Path     string
	Count    int
	Comments []ViewComment
}

// SevCount holds severity counts for the filter buttons.
type SevCount struct {
	Key   string
	Label string
	Count int
}

// ReportData is the root template data.
type ReportData struct {
	Title           string
	Lang            string
	FilesReviewed   int
	TotalComments   int
	TotalTokens     int64
	Elapsed         string
	SeverityCounts  []SevCount
	FileGroups      []FileGroup
}

// Generate creates an HTML report from review comments and writes to path.
func Generate(path string, comments []model.LlmComment, filesReviewed int, totalTokens int64, duration time.Duration) error {
	// Single pass: resolve severity once, count, and group by file
	sevCounts := make(map[string]int)
	fileMap := make(map[string][]ViewComment)

	for _, c := range comments {
		sev := c.Severity
		if sev == "" {
			sev = model.KeywordSeverity(c.Content)
		}
		// Validate severity against known constants
		if _, valid := severityLabel[sev]; !valid {
			sev = model.SeverityDefault
		}
		sevCounts[sev]++

		label := severityLabel[sev]
		if label == "" {
			label = sev
		}
		cat := categoryLabel[strings.ToLower(c.Category)]
		if cat == "" {
			cat = c.Category
		}

		lines := ""
		if c.StartLine > 0 {
			if c.EndLine > c.StartLine {
				lines = formatInt(c.StartLine) + "-" + formatInt(c.EndLine)
			} else {
				lines = formatInt(c.StartLine)
			}
		}

		vc := ViewComment{
			Severity:       sev,
			SevLabel:       label,
			Category:       cat,
			Content:        c.Content, // html/template auto-escapes
			Lines:          lines,
			SuggestionHTML: codeToHTML(c.SuggestionCode, "add"),
			ExistingHTML:   codeToHTML(c.ExistingCode, "del"),
		}

		p := c.Path
		if p == "" {
			p = "(unknown)"
		}
		fileMap[p] = append(fileMap[p], vc)
	}

	// Build sorted file groups (by path ascending)
	var paths []string
	for p := range fileMap {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	fileGroups := make([]FileGroup, 0, len(paths))
	for i, p := range paths {
		vcs := fileMap[p]
		fileGroups = append(fileGroups, FileGroup{
			ID:       formatInt(i + 1),
			Path:     p, // html/template auto-escapes
			Count:    len(vcs),
			Comments: vcs,
		})
	}

	// Build severity counts for filter buttons (ordered: CRITICAL, HIGH, MEDIUM, LOW)
	sevOrder := []string{model.SeverityCritical, model.SeverityHigh, model.SeverityMedium, model.SeverityLow}
	var sc []SevCount
	for _, k := range sevOrder {
		if c, ok := sevCounts[k]; ok && c > 0 {
			sc = append(sc, SevCount{
				Key:   k,
				Label: severityLabel[k],
				Count: c,
			})
		}
	}

	data := ReportData{
		Title:           "NoiseCheck 审查报告",
		Lang:            "zh-CN",
		FilesReviewed:   filesReviewed,
		TotalComments:   len(comments),
		TotalTokens:     totalTokens,
		Elapsed:         duration.Round(time.Second).String(),
		SeverityCounts:  sc,
		FileGroups:      fileGroups,
	}

	// Parse and execute template
	tmplContent, err := reportTemplate.ReadFile("template.html")
	if err != nil {
		return err
	}

	tmpl, err := template.New("report").Parse(string(tmplContent))
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, data)
	if err != nil {
		return err
	}

	// Sync to disk to ensure the file is fully flushed
	return f.Sync()
}

// codeToHTML wraps code in syntax-colored HTML spans.
func codeToHTML(code, mode string) template.HTML {
	if code == "" {
		return ""
	}
	escaped := html.EscapeString(code)
	klass := ""
	if mode == "add" {
		klass = ` class="code-add"`
	} else if mode == "del" {
		klass = ` class="code-del"`
	}
	return template.HTML(`<span` + klass + `>` + escaped + `</span>`)
}

// formatInt is a simple int formatter (avoids importing strconv for small numbers).
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
