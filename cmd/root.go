package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/canyouhack/steg-cli/pkg/deps"
	"github.com/canyouhack/steg-cli/pkg/detector"
	"github.com/canyouhack/steg-cli/pkg/output"
	"github.com/canyouhack/steg-cli/pkg/runner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	password  string
	skipTools []string
	onlyTools []string
	outputDir string
	verbose   bool
	noInstall bool
	timeout   int
)

var rootCmd = &cobra.Command{
	Use:   "steg",
	Short: "🔍 CanYouHack Steg — Professional Steganography Analysis Tool",
	Long: `CanYouHack Steg is a comprehensive CLI steganography analysis tool.
It automatically detects file types and runs all relevant steganography tools
concurrently, displaying results in a beautiful, organized format.

Supports: PNG, JPG, BMP, GIF, TIFF, WEBP, WAV, MP3, FLAC, OGG, AU
Tools: exiftool, binwalk, foremost, steghide, outguess, zsteg, pngcheck,
       stegoveritas, stegseek, openstego, jsteg, sox, wavsteg, and more.`,
}

var scanCmd = &cobra.Command{
	Use:   "scan <file>",
	Short: "Scan a file for hidden steganographic data",
	Long:  "Run all applicable steganography tools against the given file and display results.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		output.PrintBanner()

		// Detect file type
		fileInfo, err := detector.Detect(filePath)
		if err != nil {
			color.New(color.FgRed, color.Bold).Printf("\n  ❌ Error: %v\n\n", err)
			os.Exit(1)
		}

		output.PrintFileInfo(fileInfo.Name, fileInfo.Path, fileInfo.MimeType, fileInfo.Size, string(fileInfo.Category))

		if fileInfo.Category == detector.CategoryUnknown {
			color.New(color.FgYellow, color.Bold).Println("  ⚠  Unknown file type. Running general tools only.")
			fmt.Println()
		}

		// Check for missing tools and notify
		toolStatus := deps.CheckAll()
		var missing []string
		for name, available := range toolStatus {
			if !available {
				missing = append(missing, name)
			}
		}
		output.PrintDepsNotice(missing)

		// Run scan
		opts := runner.RunOpts{
			Password:  password,
			Verbose:   verbose,
			OutputDir: outputDir,
			Skip:      skipTools,
			Only:      onlyTools,
			Timeout:   60 * 1e9, // 60s in nanoseconds
		}

		if timeout > 0 {
			opts.Timeout = runner.RunOpts{}.Timeout // use default
		}

		scanResult := runner.RunAll(fileInfo, opts)

		// Clear progress line
		fmt.Printf("\r%s\r", strings.Repeat(" ", 80))

		// Print results
		runner.PrintResults(scanResult)

		// Show output directory info if files were extracted
		if outputDir != "" {
			cyan := color.New(color.FgCyan)
			cyan.Printf("  📁 Extracted files saved to: %s\n\n", outputDir)
		} else {
			cyan := color.New(color.FgCyan)
			cyan.Printf("  📁 Extracted files saved to: /tmp/steg-cli-output/\n\n")
		}
	},
}

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Check status of all steganography tools",
	Long:  "Display the installation status of all supported steganography tools.",
	Run: func(cmd *cobra.Command, args []string) {
		output.PrintBanner()
		deps.PrintStatus()
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all missing steganography tools",
	Long:  "Automatically install all missing steganography tools using the system's package manager.",
	Run: func(cmd *cobra.Command, args []string) {
		output.PrintBanner()
		deps.InstallMissing()
	},
}

func init() {
	scanCmd.Flags().StringVarP(&password, "password", "p", "", "Password for steghide/openstego extraction")
	scanCmd.Flags().StringSliceVar(&skipTools, "skip", nil, "Skip specific tools (comma-separated)")
	scanCmd.Flags().StringSliceVar(&onlyTools, "only", nil, "Run only specific tools (comma-separated)")
	scanCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for extracted files (default: /tmp/steg-cli-output)")
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	scanCmd.Flags().IntVarP(&timeout, "timeout", "t", 60, "Timeout per tool in seconds")

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(depsCmd)
	rootCmd.AddCommand(installCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
