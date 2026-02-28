package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	banner = `
   ██████╗ █████╗ ███╗   ██╗██╗   ██╗ ██████╗ ██╗   ██╗
  ██╔════╝██╔══██╗████╗  ██║╚██╗ ██╔╝██╔═══██╗██║   ██║
  ██║     ███████║██╔██╗ ██║ ╚████╔╝ ██║   ██║██║   ██║
  ██║     ██╔══██║██║╚██╗██║  ╚██╔╝  ██║   ██║██║   ██║
  ╚██████╗██║  ██║██║ ╚████║   ██║   ╚██████╔╝╚██████╔╝
   ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝  ╚═════╝
         ██╗  ██╗ █████╗  ██████╗██╗  ██╗
         ██║  ██║██╔══██╗██╔════╝██║ ██╔╝
         ███████║███████║██║     █████╔╝
         ██╔══██║██╔══██║██║     ██╔═██╗
         ██║  ██║██║  ██║╚██████╗██║  ██╗
         ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
              ╔═╗╔╦╗╔═╗╔═╗
              ╚═╗ ║ ║╣ ║ ╦
              ╚═╝ ╩ ╚═╝╚═╝  v1.0.0
`
)

// Result represents the output of a single tool run
type Result struct {
	ToolName  string
	Category  string
	Output    string
	Error     error
	Duration  time.Duration
	Skipped   bool
	SkipReason string
}

// PrintBanner prints the ASCII art banner
func PrintBanner() {
	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Println(banner)
}

// PrintFileInfo prints file analysis header
func PrintFileInfo(name, path, mime string, size int64, category string) {
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite)
	yellow := color.New(color.FgYellow, color.Bold)

	cyan.Println("  ╔══════════════════════════════════════════════════════════════╗")
	cyan.Printf("  ║")
	yellow.Printf("  📄 TARGET FILE")
	cyan.Println("                                              ║")
	cyan.Println("  ╠══════════════════════════════════════════════════════════════╣")

	printField := func(label, value string) {
		cyan.Printf("  ║  ")
		white.Printf("%-12s", label)
		fmt.Printf("%-48s", value)
		cyan.Println("║")
	}

	printField("File:", name)
	printField("Path:", truncate(path, 47))
	printField("MIME:", mime)
	printField("Size:", formatSize(size))
	printField("Category:", strings.ToUpper(category))

	cyan.Println("  ╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// PrintScanStart prints the scan start message
func PrintScanStart(toolCount int) {
	yellow := color.New(color.FgYellow, color.Bold)
	yellow.Printf("  🔍 Starting scan with %d tools...\n\n", toolCount)
}

// PrintCategoryHeader prints a category separator
func PrintCategoryHeader(category string) {
	magenta := color.New(color.FgMagenta, color.Bold)
	fmt.Println()
	magenta.Printf("  ┌─── %s ", strings.ToUpper(category))
	magenta.Printf("%s\n", strings.Repeat("─", 55-len(category)))
}

// PrintToolResult prints a single tool's result
func PrintToolResult(result *Result) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)
	gray := color.New(color.FgHiBlack)
	white := color.New(color.FgWhite)

	if result.Skipped {
		gray.Printf("  │ ⊘ %-18s", result.ToolName)
		gray.Printf("skipped: %s\n", result.SkipReason)
		return
	}

	if result.Error != nil {
		red.Printf("  │ ✗ %-18s", result.ToolName)
		gray.Printf("(%s) ", result.Duration.Round(time.Millisecond))
		red.Printf("error: %v\n", result.Error)
		return
	}

	output := strings.TrimSpace(result.Output)
	if output == "" {
		gray.Printf("  │ ○ %-18s", result.ToolName)
		gray.Printf("(%s) ", result.Duration.Round(time.Millisecond))
		gray.Println("no output")
		return
	}

	green.Printf("  │ ✓ %-18s", result.ToolName)
	cyan.Printf("(%s)\n", result.Duration.Round(time.Millisecond))

	lines := strings.Split(output, "\n")
	maxLines := 30

	for i, line := range lines {
		if i >= maxLines {
			yellow.Printf("  │   ... and %d more lines\n", len(lines)-maxLines)
			break
		}
		white.Printf("  │   %s\n", line)
	}
}

// PrintSummary prints the final summary
func PrintSummary(results []*Result, totalDuration time.Duration) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)

	total := len(results)
	success := 0
	failed := 0
	skipped := 0
	withOutput := 0

	for _, r := range results {
		if r.Skipped {
			skipped++
		} else if r.Error != nil {
			failed++
		} else {
			success++
			if strings.TrimSpace(r.Output) != "" {
				withOutput++
			}
		}
	}

	const boxW = 60 // inner width between ║...║

	fmt.Println()
	cyan.Printf("  ╔%s╗\n", strings.Repeat("═", boxW))
	header := "  📊 SCAN SUMMARY"
	cyan.Printf("  ║%-*s║\n", boxW, header)
	cyan.Printf("  ╠%s╣\n", strings.Repeat("═", boxW))

	printStat := func(icon, label string, value string, c *color.Color) {
		// icon(emoji ~2 display cols) + space + label + value, pad to boxW
		content := fmt.Sprintf("  %s %-18s%s", icon, label, value)
		pad := boxW - len(content)
		if pad < 0 {
			pad = 0
		}
		cyan.Printf("  ║")
		c.Printf("%s%s", content, strings.Repeat(" ", pad))
		cyan.Printf("║\n")
	}

	printStat("📋", "Total tools:", fmt.Sprintf("%d", total), cyan)
	printStat("✅", "Successful:", fmt.Sprintf("%d", success), green)
	printStat("📝", "With output:", fmt.Sprintf("%d", withOutput), yellow)
	printStat("❌", "Failed:", fmt.Sprintf("%d", failed), red)
	printStat("⊘ ", "Skipped:", fmt.Sprintf("%d", skipped), yellow)
	printStat("⏱ ", "Duration:", totalDuration.Round(time.Millisecond).String(), cyan)

	cyan.Printf("  ╚%s╝\n", strings.Repeat("═", boxW))
	fmt.Println()
}

// PrintProgress prints a progress update during scanning
func PrintProgress(toolName string, current, total int) {
	cyan := color.New(color.FgCyan)
	cyan.Printf("\r  ⏳ [%d/%d] Running %s...                    ", current, total, toolName)
}

// PrintDepsNotice prints notice about missing tools
func PrintDepsNotice(missing []string) {
	yellow := color.New(color.FgYellow)
	if len(missing) > 0 {
		yellow.Printf("  ⚠  %d tools not installed (will be skipped): %s\n", len(missing), strings.Join(missing, ", "))
		yellow.Println("  💡 Run 'steg install' to install all missing tools.")
		fmt.Println()
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "..." + s[len(s)-max+3:]
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
