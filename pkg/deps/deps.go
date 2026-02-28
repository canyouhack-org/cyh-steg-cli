package deps

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

// Distro represents the Linux distribution
type Distro string

const (
	DistroDebian  Distro = "debian"  // Ubuntu, Debian, Kali, Mint
	DistroArch    Distro = "arch"    // Arch, Manjaro, EndeavourOS
	DistroFedora  Distro = "fedora"  // Fedora, RHEL, CentOS, Rocky
	DistroSuse    Distro = "suse"    // openSUSE
	DistroUnknown Distro = "unknown"
)

// ToolDep represents a tool dependency
type ToolDep struct {
	Name        string
	Binary      string   // binary name to check
	AltBinaries []string // alternative binary names
	InstallType string   // "apt", "pip", "gem", "go", "manual", "builtin"
	AptPkg      string
	PacmanPkg   string
	AurPkg      string   // AUR package name (for Arch when not in official repos)
	DnfPkg      string
	ZypperPkg   string
	PipPkg      string
	GemPkg      string
	GoPath      string
	ManualURL   string
	Description string
}

// AllTools returns the complete list of tool dependencies
func AllTools() []ToolDep {
	return []ToolDep{
		{
			Name: "file", Binary: "file",
			InstallType: "builtin",
			Description: "File type identification",
		},
		{
			Name: "strings", Binary: "strings",
			InstallType: "builtin",
			Description: "Extract printable strings",
		},
		{
			Name: "xxd", Binary: "xxd",
			InstallType: "system",
			AptPkg: "xxd", PacmanPkg: "xxd", DnfPkg: "vim-common", ZypperPkg: "vim-data-common",
			Description: "Hex dump utility",
		},
		{
			Name: "exiftool", Binary: "exiftool",
			InstallType: "system",
			AptPkg: "libimage-exiftool-perl", PacmanPkg: "perl-image-exiftool",
			DnfPkg: "perl-Image-ExifTool", ZypperPkg: "exiftool",
			Description: "Metadata extraction",
		},
		{
			Name: "binwalk", Binary: "binwalk",
			InstallType: "system",
			AptPkg: "binwalk", PacmanPkg: "binwalk", DnfPkg: "binwalk", ZypperPkg: "binwalk",
			Description: "Embedded file detection",
		},
		{
			Name: "foremost", Binary: "foremost",
			InstallType: "system",
			AptPkg: "foremost", PacmanPkg: "foremost", DnfPkg: "foremost", ZypperPkg: "foremost",
			Description: "File carving tool",
		},
		{
			Name: "steghide", Binary: "steghide",
			InstallType: "system",
			AptPkg: "steghide", PacmanPkg: "steghide", DnfPkg: "steghide", ZypperPkg: "steghide",
			Description: "Steganography hide/extract (JPG/BMP/WAV/AU)",
		},
		{
			Name: "zsteg", Binary: "zsteg",
			InstallType: "gem",
			GemPkg: "zsteg",
			Description: "LSB steganography (PNG/BMP)",
		},
		{
			Name: "pngcheck", Binary: "pngcheck",
			InstallType: "system",
			AptPkg: "pngcheck", PacmanPkg: "pngcheck", DnfPkg: "pngcheck", ZypperPkg: "pngcheck",
			Description: "PNG integrity check",
		},
		{
			Name: "stegoveritas", Binary: "stegoveritas",
			InstallType: "pip",
			PipPkg: "stegoveritas",
			Description: "Advanced image steganalysis",
		},
		{
			Name: "stegseek", Binary: "stegseek",
			InstallType: "system",
			AptPkg: "stegseek", PacmanPkg: "", AurPkg: "stegseek",
			DnfPkg: "", ZypperPkg: "",
			ManualURL: "https://github.com/RickdeJager/stegseek/releases",
			Description: "Fast steghide brute-force cracker",
		},
		{
			Name: "openstego", Binary: "openstego",
			InstallType: "system",
			AptPkg: "openstego", PacmanPkg: "", AurPkg: "openstego",
			DnfPkg: "", ZypperPkg: "",
			ManualURL: "https://github.com/syvaidya/OpenStego/releases",
			Description: "OpenStego extraction (PNG)",
		},
		{
			Name: "jsteg", Binary: "jsteg",
			InstallType: "go",
			GoPath: "lukechampine.com/jsteg@latest",
			Description: "JPEG steganography (no password)",
		},
		{
			Name: "GraphicsMagick", Binary: "gm",
			InstallType: "system",
			AptPkg: "graphicsmagick", PacmanPkg: "graphicsmagick",
			DnfPkg: "GraphicsMagick", ZypperPkg: "GraphicsMagick",
			Description: "Image identification and analysis",
		},
		{
			Name: "sox", Binary: "sox",
			InstallType: "system",
			AptPkg: "sox", PacmanPkg: "sox", DnfPkg: "sox", ZypperPkg: "sox",
			Description: "Audio spectrogram generation",
		},
		{
			Name: "stegolsb", Binary: "stegolsb",
			InstallType: "pip",
			PipPkg: "stego-lsb",
			Description: "WAV LSB steganography",
		},
		{
			Name: "stegsolve", Binary: "python3",
			InstallType: "pip",
			PipPkg: "Pillow",
			Description: "Stegsolve-like bitplane extraction",
		},
	}
}

// DetectDistro detects the current Linux distribution
func DetectDistro() Distro {
	if runtime.GOOS != "linux" {
		return DistroUnknown
	}

	// Check for each package manager
	if isCommandAvailable("apt-get") {
		return DistroDebian
	}
	if isCommandAvailable("pacman") {
		return DistroArch
	}
	if isCommandAvailable("dnf") {
		return DistroFedora
	}
	if isCommandAvailable("zypper") {
		return DistroSuse
	}

	// Fallback: read /etc/os-release
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "ubuntu") || strings.Contains(content, "debian") || strings.Contains(content, "kali") || strings.Contains(content, "mint") {
			return DistroDebian
		}
		if strings.Contains(content, "arch") || strings.Contains(content, "manjaro") || strings.Contains(content, "endeavour") {
			return DistroArch
		}
		if strings.Contains(content, "fedora") || strings.Contains(content, "rhel") || strings.Contains(content, "centos") || strings.Contains(content, "rocky") {
			return DistroFedora
		}
		if strings.Contains(content, "suse") {
			return DistroSuse
		}
	}

	return DistroUnknown
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// IsToolAvailable checks if a specific tool is available
func IsToolAvailable(binary string) bool {
	return isCommandAvailable(binary)
}

// CheckAll checks all tools and returns their status
func CheckAll() map[string]bool {
	status := make(map[string]bool)
	for _, tool := range AllTools() {
		available := isCommandAvailable(tool.Binary)
		if !available {
			for _, alt := range tool.AltBinaries {
				if isCommandAvailable(alt) {
					available = true
					break
				}
			}
		}
		status[tool.Name] = available
	}
	return status
}

// PrintStatus prints the status of all tools with color
func PrintStatus() {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite)

	distro := DetectDistro()
	cyan.Printf("\n  🖥  Detected System: ")
	white.Printf("%s\n\n", distro)

	tools := AllTools()
	installed := 0

	fmt.Printf("  %-20s %-12s %s\n", "TOOL", "STATUS", "DESCRIPTION")
	fmt.Printf("  %s\n", strings.Repeat("─", 65))

	for _, tool := range tools {
		available := isCommandAvailable(tool.Binary)
		if !available {
			for _, alt := range tool.AltBinaries {
				if isCommandAvailable(alt) {
					available = true
					break
				}
			}
		}

		if available {
			installed++
			fmt.Printf("  %-20s ", tool.Name)
			green.Printf("%-12s", "✓ ready")
			fmt.Printf(" %s\n", tool.Description)
		} else {
			fmt.Printf("  %-20s ", tool.Name)
			red.Printf("%-12s", "✗ missing")
			fmt.Printf(" %s\n", tool.Description)
		}
	}

	fmt.Printf("\n  %s\n", strings.Repeat("─", 65))
	if installed == len(tools) {
		green.Printf("  ✓ All %d tools are installed and ready!\n\n", installed)
	} else {
		yellow.Printf("  ⚠ %d/%d tools installed. Missing tools will be skipped during scan.\n", installed, len(tools))
		yellow.Printf("  Run 'steg install' to install missing tools.\n\n")
	}
}

// InstallMissing attempts to install all missing tools
func InstallMissing() error {
	distro := DetectDistro()
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow)

	cyan.Printf("\n  📦 Installing missing dependencies...\n")
	cyan.Printf("  🖥  Detected distro: %s\n\n", distro)

	tools := AllTools()
	for _, tool := range tools {
		if isCommandAvailable(tool.Binary) {
			green.Printf("  ✓ %s already installed\n", tool.Name)
			continue
		}

		if tool.InstallType == "builtin" {
			continue
		}

		fmt.Printf("  ⏳ Installing %s...", tool.Name)

		var err error
		switch tool.InstallType {
		case "system":
			err = installSystem(tool, distro)
		case "pip":
			err = installPip(tool)
		case "gem":
			err = installGem(tool)
		case "go":
			err = installGo(tool)
		}

		if err != nil {
			red.Printf(" ✗ failed: %v\n", err)
			if tool.ManualURL != "" {
				yellow.Printf("    → Manual install: %s\n", tool.ManualURL)
			}
		} else {
			green.Printf(" ✓ done\n")
		}
	}

	// Ensure rockyou.txt exists for brute-force tools
	ensureRockyou()

	green.Printf("\n  ✓ Installation complete!\n\n")
	return nil
}

func installSystem(tool ToolDep, distro Distro) error {
	var cmd *exec.Cmd
	switch distro {
	case DistroDebian:
		if tool.AptPkg == "" {
			return fmt.Errorf("no apt package available")
		}
		cmd = exec.Command("sudo", "apt-get", "install", "-y", tool.AptPkg)
	case DistroArch:
		// Try official repo first, then AUR
		if tool.PacmanPkg != "" {
			cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", tool.PacmanPkg)
		} else if tool.AurPkg != "" {
			return installAUR(tool.AurPkg)
		} else {
			return fmt.Errorf("no pacman/AUR package available")
		}
	case DistroFedora:
		if tool.DnfPkg == "" {
			return fmt.Errorf("no dnf package available")
		}
		cmd = exec.Command("sudo", "dnf", "install", "-y", tool.DnfPkg)
	case DistroSuse:
		if tool.ZypperPkg == "" {
			return fmt.Errorf("no zypper package available")
		}
		cmd = exec.Command("sudo", "zypper", "install", "-y", tool.ZypperPkg)
	default:
		return fmt.Errorf("unknown distro, cannot install automatically")
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	// If pacman fails and AUR package exists, try AUR
	if err != nil && distro == DistroArch && tool.AurPkg != "" {
		return installAUR(tool.AurPkg)
	}
	return err
}

// installAUR tries to install a package from AUR using yay or paru
func installAUR(pkg string) error {
	// Try yay first (yay handles sudo internally, do NOT run with sudo)
	if isCommandAvailable("yay") {
		cmd := exec.Command("yay", "-S", "--noconfirm", "--needed", pkg)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	// Try paru (paru also handles sudo internally)
	if isCommandAvailable("paru") {
		cmd := exec.Command("paru", "-S", "--noconfirm", "--needed", pkg)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no AUR helper (yay/paru) found or install failed for %s", pkg)
}

func installPip(tool ToolDep) error {
	// Try pip3 first, then pip
	pip := "pip3"
	if !isCommandAvailable("pip3") {
		pip = "pip"
	}
	if !isCommandAvailable(pip) {
		return fmt.Errorf("pip/pip3 not found, install python3-pip first")
	}
	// Use --break-system-packages for modern Python (3.11+)
	cmd := exec.Command(pip, "install", "--break-system-packages", tool.PipPkg)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	if err != nil {
		// Fallback without --break-system-packages for older pip
		cmd2 := exec.Command(pip, "install", tool.PipPkg)
		cmd2.Stdout = io.Discard
		cmd2.Stderr = io.Discard
		return cmd2.Run()
	}
	return nil
}

func installGem(tool ToolDep) error {
	if !isCommandAvailable("gem") {
		return fmt.Errorf("gem not found, install ruby first")
	}
	cmd := exec.Command("gem", "install", tool.GemPkg)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func installGo(tool ToolDep) error {
	if !isCommandAvailable("go") {
		return fmt.Errorf("go not found")
	}
	cmd := exec.Command("go", "install", tool.GoPath)
	// Ensure GOBIN is in PATH-accessible location
	home, _ := os.UserHomeDir()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(home, "go")
	}
	cmd.Env = append(os.Environ(), "GOBIN="+filepath.Join(gopath, "bin"))
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

// RockyouPath returns the path to rockyou.txt
func RockyouPath() string {
	// Check common locations
	paths := []string{
		"/usr/share/wordlists/rockyou.txt",
		"/usr/share/seclists/Passwords/Leaked-Databases/rockyou.txt",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Check our local wordlist dir
	home, _ := os.UserHomeDir()
	localPath := filepath.Join(home, ".steg-cli", "wordlists", "rockyou.txt")
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}

	return localPath // Return the local path even if not downloaded yet
}

// EnsureRockyouExists checks if rockyou.txt exists and downloads if needed
func EnsureRockyouExists() string {
	return ensureRockyou()
}

func ensureRockyou() string {
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	// Check system locations first
	systemPaths := []string{
		"/usr/share/wordlists/rockyou.txt",
		"/usr/share/seclists/Passwords/Leaked-Databases/rockyou.txt",
	}
	for _, p := range systemPaths {
		if _, err := os.Stat(p); err == nil {
			green.Printf("  ✓ rockyou.txt found at %s\n", p)
			return p
		}
	}

	// Check compressed versions
	gzPaths := []string{
		"/usr/share/wordlists/rockyou.txt.gz",
	}
	for _, p := range gzPaths {
		if _, err := os.Stat(p); err == nil {
			yellow.Printf("  📦 Found compressed rockyou at %s, extracting...\n", p)
			cmd := exec.Command("sudo", "gzip", "-dk", p)
			if err := cmd.Run(); err == nil {
				txtPath := strings.TrimSuffix(p, ".gz")
				green.Printf("  ✓ Extracted to %s\n", txtPath)
				return txtPath
			}
		}
	}

	// Download to local dir
	home, _ := os.UserHomeDir()
	localDir := filepath.Join(home, ".steg-cli", "wordlists")
	localPath := filepath.Join(localDir, "rockyou.txt")

	if _, err := os.Stat(localPath); err == nil {
		green.Printf("  ✓ rockyou.txt found at %s\n", localPath)
		return localPath
	}

	yellow.Printf("  ⏳ Downloading rockyou.txt (14MB)...\n")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		yellow.Printf("  ⚠ Cannot create wordlist directory: %v\n", err)
		return ""
	}

	// Download from GitHub
	url := "https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt"
	resp, err := http.Get(url)
	if err != nil {
		yellow.Printf("  ⚠ Cannot download rockyou.txt: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	out, err := os.Create(localPath)
	if err != nil {
		yellow.Printf("  ⚠ Cannot create file: %v\n", err)
		return ""
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		yellow.Printf("  ⚠ Download failed: %v\n", err)
		os.Remove(localPath)
		return ""
	}

	green.Printf("  ✓ Downloaded rockyou.txt to %s\n", localPath)
	return localPath
}
