package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/canyouhack/steg-cli/pkg/deps"
	"github.com/canyouhack/steg-cli/pkg/detector"
	"github.com/canyouhack/steg-cli/pkg/output"
)

// RunOpts contains options for running tools
type RunOpts struct {
	Password  string
	Verbose   bool
	OutputDir string
	Skip      []string
	Only      []string
	Timeout   time.Duration
}

// Tool represents a steganography analysis tool
type Tool struct {
	Name           string
	Binary         string
	Category       string // "general", "image", "audio"
	SupportedTypes []detector.FileCategory
	BuildCmd       func(filePath string, opts RunOpts) *exec.Cmd
}

// ScanResult holds all results from a scan
type ScanResult struct {
	FileInfo *detector.FileInfo
	Results  []*output.Result
	Duration time.Duration
}

// RunAll runs all applicable tools concurrently on the given file
func RunAll(fileInfo *detector.FileInfo, opts RunOpts) *ScanResult {
	if opts.Timeout == 0 {
		opts.Timeout = 60 * time.Second
	}

	allTools := GetAllTools(opts)
	applicable := filterTools(allTools, fileInfo, opts)

	output.PrintScanStart(len(applicable))

	results := make([]*output.Result, len(applicable))
	var wg sync.WaitGroup

	startTime := time.Now()

	for i, tool := range applicable {
		wg.Add(1)
		go func(idx int, t *Tool) {
			defer wg.Done()

			result := &output.Result{
				ToolName: t.Name,
				Category: t.Category,
			}

			// Check if the binary exists
			if !deps.IsToolAvailable(t.Binary) {
				result.Skipped = true
				result.SkipReason = fmt.Sprintf("%s not installed", t.Binary)
				results[idx] = result
				return
			}

			// Build the command
			cmd := t.BuildCmd(fileInfo.Path, opts)
			if cmd == nil {
				result.Skipped = true
				result.SkipReason = "command not applicable"
				results[idx] = result
				return
			}

			// Run with timeout
			ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
			defer cancel()

			cmd2 := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
			cmd2.Dir = cmd.Dir
			cmd2.Env = cmd.Env

			toolStart := time.Now()
			out, err := cmd2.CombinedOutput()
			result.Duration = time.Since(toolStart)

			if ctx.Err() == context.DeadlineExceeded {
				result.Error = fmt.Errorf("timeout after %s", opts.Timeout)
			} else if err != nil {
				// Some tools return non-zero exit codes even on success
				outStr := strings.TrimSpace(string(out))
				if outStr != "" {
					result.Output = outStr
				} else {
					result.Error = fmt.Errorf("%v", err)
				}
			} else {
				result.Output = strings.TrimSpace(string(out))
			}

			results[idx] = result
		}(i, tool)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	return &ScanResult{
		FileInfo: fileInfo,
		Results:  results,
		Duration: totalDuration,
	}
}

// filterTools returns only tools that are applicable to the given file type
func filterTools(tools []*Tool, fileInfo *detector.FileInfo, opts RunOpts) []*Tool {
	var filtered []*Tool

	skipSet := make(map[string]bool)
	for _, s := range opts.Skip {
		skipSet[strings.ToLower(s)] = true
	}

	onlySet := make(map[string]bool)
	for _, o := range opts.Only {
		onlySet[strings.ToLower(o)] = true
	}

	for _, tool := range tools {
		// Check --skip
		if skipSet[strings.ToLower(tool.Name)] {
			continue
		}

		// Check --only
		if len(onlySet) > 0 && !onlySet[strings.ToLower(tool.Name)] {
			continue
		}

		// Check supported types
		if len(tool.SupportedTypes) > 0 {
			supported := false
			for _, st := range tool.SupportedTypes {
				if st == fileInfo.Category {
					supported = true
					break
				}
			}
			if !supported {
				continue
			}
		}

		filtered = append(filtered, tool)
	}

	return filtered
}

// PrintResults prints all results grouped by category
func PrintResults(scanResult *ScanResult) {
	categories := []string{"general", "image", "audio", "text"}
	categoryNames := map[string]string{
		"general": "🔧 General Analysis",
		"image":   "🖼️  Image Steganography",
		"audio":   "🎵 Audio Steganography",
		"text":    "📝 Text / Misc Steganography",
	}

	for _, cat := range categories {
		var catResults []*output.Result
		for _, r := range scanResult.Results {
			if r != nil && r.Category == cat {
				catResults = append(catResults, r)
			}
		}

		if len(catResults) == 0 {
			continue
		}

		output.PrintCategoryHeader(categoryNames[cat])
		for _, r := range catResults {
			output.PrintToolResult(r)
		}
	}

	output.PrintSummary(scanResult.Results, scanResult.Duration)
}
