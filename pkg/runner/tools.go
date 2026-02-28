package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/canyouhack/steg-cli/pkg/deps"
	"github.com/canyouhack/steg-cli/pkg/detector"
)

// allImageTypes is a convenience for tools that support most image types
var allImageTypes = []detector.FileCategory{
	detector.CategoryPNG, detector.CategoryJPG, detector.CategoryBMP,
	detector.CategoryGIF, detector.CategoryTIFF, detector.CategoryWEBP,
}

// allAudioTypes is a convenience for tools that support most audio types
var allAudioTypes = []detector.FileCategory{
	detector.CategoryWAV, detector.CategoryMP3, detector.CategoryFLAC,
	detector.CategoryOGG, detector.CategoryAU,
}

// allTypes combines all supported types
var allTypes = append(append([]detector.FileCategory{}, allImageTypes...), allAudioTypes...)

// GetAllTools returns all available tool definitions
func GetAllTools(opts RunOpts) []*Tool {
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "/tmp/steg-cli-output"
	}
	os.MkdirAll(outputDir, 0755)

	return []*Tool{
		// ========================
		// GENERAL TOOLS
		// ========================
		{
			Name:           "file",
			Binary:         "file",
			Category:       "general",
			SupportedTypes: allTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("file", "-b", "--mime", fp)
			},
		},
		{
			Name:           "exiftool",
			Binary:         "exiftool",
			Category:       "general",
			SupportedTypes: allTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("exiftool", fp)
			},
		},
		{
			Name:           "binwalk",
			Binary:         "binwalk",
			Category:       "general",
			SupportedTypes: allTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("binwalk", fp)
			},
		},
		{
			Name:           "strings",
			Binary:         "strings",
			Category:       "general",
			SupportedTypes: allTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("strings", "-n", "8", fp)
			},
		},
		{
			Name:           "hexdump",
			Binary:         "xxd",
			Category:       "general",
			SupportedTypes: allTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("bash", "-c", fmt.Sprintf("xxd '%s' | head -50", fp))
			},
		},
		{
			Name:           "foremost",
			Binary:         "foremost",
			Category:       "general",
			SupportedTypes: allTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				outDir := filepath.Join(outputDir, "foremost")
				os.RemoveAll(outDir)
				return exec.Command("foremost", "-t", "all", "-i", fp, "-o", outDir)
			},
		},

		// ========================
		// IMAGE TOOLS
		// ========================
		{
			Name:     "zsteg",
			Binary:   "zsteg",
			Category: "image",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryPNG, detector.CategoryBMP,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("zsteg", fp, "--all")
			},
		},
		{
			Name:     "steghide-extract",
			Binary:   "steghide",
			Category: "image",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryJPG, detector.CategoryBMP,
				detector.CategoryWAV, detector.CategoryAU,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				pass := opts.Password
				if pass == "" {
					pass = ""
				}
				outFile := filepath.Join(outputDir, "steghide_extracted.txt")
				return exec.Command("steghide", "extract", "-sf", fp, "-p", pass, "-xf", outFile, "-f")
			},
		},
		{
			Name:     "steghide-info",
			Binary:   "steghide",
			Category: "image",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryJPG, detector.CategoryBMP,
				detector.CategoryWAV, detector.CategoryAU,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				// Use -p "" to auto-answer passphrase prompt
				return exec.Command("steghide", "info", fp, "-p", "")
			},
		},
		{
			Name:     "pngcheck",
			Binary:   "pngcheck",
			Category: "image",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryPNG,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("pngcheck", "-vtp", fp)
			},
		},
		{
			Name:           "identify",
			Binary:         "gm",
			Category:       "image",
			SupportedTypes: allImageTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("gm", "identify", "-verbose", fp)
			},
		},
		{
			Name:     "jsteg",
			Binary:   "jsteg",
			Category: "image",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryJPG,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("jsteg", "reveal", fp)
			},
		},
		{
			Name:     "openstego",
			Binary:   "openstego",
			Category: "image",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryPNG,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				outFile := filepath.Join(outputDir, "openstego_extracted")
				pass := opts.Password
				args := []string{"extract", "--algorithm", "RandomLSB", "-sf", fp, "-xd", outFile}
				if pass != "" {
					args = append(args, "-p", pass)
				}
				return exec.Command("openstego", args...)
			},
		},
		{
			Name:     "stegoveritas",
			Binary:   "stegoveritas",
			Category: "image",
			SupportedTypes: allImageTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				outDir := filepath.Join(outputDir, "stegoveritas")
				os.RemoveAll(outDir)
				os.MkdirAll(outDir, 0755)
				return exec.Command("stegoveritas", "-out", outDir, fp)
			},
		},
		{
			Name:     "stegseek",
			Binary:   "stegseek",
			Category: "image",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryJPG, detector.CategoryBMP,
				detector.CategoryWAV, detector.CategoryAU,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				rockyou := deps.RockyouPath()
				if _, err := os.Stat(rockyou); err != nil {
					rockyou = deps.EnsureRockyouExists()
					if rockyou == "" {
						return nil
					}
				}
				outFile := filepath.Join(outputDir, "stegseek_extracted.txt")
				return exec.Command("stegseek", fp, rockyou, outFile, "--force")
			},
		},
		{
			Name:           "stegsolve",
			Binary:         "python3",
			Category:       "image",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryPNG, detector.CategoryBMP, detector.CategoryJPG,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				outDir := filepath.Join(outputDir, "stegsolve_planes")
				os.MkdirAll(outDir, 0755)
				script := fmt.Sprintf(`
import sys, os
try:
    from PIL import Image
except ImportError:
    print("Pillow not installed. Run 'pip install Pillow'")
    sys.exit(1)

out_dir = '%s'
img_path = '%s'
try:
    img = Image.open(img_path).convert('RGB')
    width, height = img.size
    
    # We will avoid loading all pixels at once in python if possible,
    # but load() is fast enough for typical steganography images.
    pixels = img.load()
    
    channels = ['Red', 'Green', 'Blue']
    for c in range(3):
        for bit in range(8):
            # Create a 1-bit image
            out = Image.new('1', (width, height))
            out_pixels = out.load()
            for y in range(height):
                for x in range(width):
                    val = pixels[x, y][c]
                    out_pixels[x, y] = (val >> bit) & 1
            
            plane_name = f"{channels[c]}_bit{bit}.png"
            out.save(os.path.join(out_dir, plane_name))
            
    print(f"Extracted 24 RGB bitplanes to {out_dir}/")
except Exception as e:
    print(f"Error extracting bitplanes: {e}")
`, outDir, fp)
				return exec.Command("python3", "-c", script)
			},
		},

		// ========================
		// AUDIO TOOLS
		// ========================
		{
			Name:     "steghide-audio",
			Binary:   "steghide",
			Category: "audio",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryWAV, detector.CategoryAU,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				pass := opts.Password
				if pass == "" {
					pass = ""
				}
				outFile := filepath.Join(outputDir, "steghide_audio_extracted.txt")
				return exec.Command("steghide", "extract", "-sf", fp, "-p", pass, "-xf", outFile, "-f")
			},
		},
		{
			Name:     "steghide-audio-info",
			Binary:   "steghide",
			Category: "audio",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryWAV, detector.CategoryAU,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				return exec.Command("steghide", "info", fp, "-p", "")
			},
		},
		{
			Name:     "wavsteg",
			Binary:   "stegolsb",
			Category: "audio",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryWAV,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				outFile := filepath.Join(outputDir, "wavsteg_extracted.txt")
				return exec.Command("stegolsb", "wavsteg", "-r", "-i", fp, "-o", outFile, "-n", "2", "-b", "1000")
			},
		},
		{
			Name:     "sox-spectrogram",
			Binary:   "sox",
			Category: "audio",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryWAV, detector.CategoryMP3, detector.CategoryFLAC,
				detector.CategoryOGG,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				outFile := filepath.Join(outputDir, "spectrogram.png")
				cmd := exec.Command("sox", fp, "-n", "spectrogram", "-o", outFile)
				return cmd
			},
		},
		{
			Name:     "stegseek-audio",
			Binary:   "stegseek",
			Category: "audio",
			SupportedTypes: []detector.FileCategory{
				detector.CategoryWAV, detector.CategoryAU,
			},
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				rockyou := deps.RockyouPath()
				if _, err := os.Stat(rockyou); err != nil {
					rockyou = deps.EnsureRockyouExists()
					if rockyou == "" {
						return nil
					}
				}
				outFile := filepath.Join(outputDir, "stegseek_audio_extracted.txt")
				return exec.Command("stegseek", fp, rockyou, outFile, "--force")
			},
		},

		// ========================
		// TEXT / MISC TOOLS
		// ========================
		{
			Name:           "unicode-steg",
			Binary:         "python3",
			Category:       "text",
			SupportedTypes: allTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				// Detect zero-width Unicode characters (U+200B, U+200C, U+200D, U+FEFF, U+200E, U+200F, U+202A-202E)
				script := fmt.Sprintf(`
import sys
zwc = {'\u200b':'ZWSP', '\u200c':'ZWNJ', '\u200d':'ZWJ', '\ufeff':'BOM/ZWNBS',
       '\u200e':'LRM', '\u200f':'RLM', '\u202a':'LRE', '\u202b':'RLE',
       '\u202c':'PDF', '\u202d':'LRO', '\u202e':'RLO', '\u2060':'WJ', '\u2061':'FA',
       '\u2062':'IT', '\u2063':'IS', '\u2064':'IP'}
try:
    with open('%s', 'r', errors='ignore') as f:
        data = f.read()
    found = {}
    bits = []
    for ch in data:
        if ch in zwc:
            name = zwc[ch]
            found[name] = found.get(name, 0) + 1
            bits.append('1' if ch in ['\u200b','\u200d','\u200e'] else '0')
    if found:
        print("[!] Zero-width Unicode characters detected!")
        for name, count in sorted(found.items(), key=lambda x: -x[1]):
            print(f"    {name}: {count} occurrences")
        total = sum(found.values())
        print(f"    Total: {total} hidden characters")
        if len(bits) >= 8:
            decoded = ''
            for i in range(0, min(len(bits), 800), 8):
                byte_bits = ''.join(bits[i:i+8])
                if len(byte_bits) == 8:
                    ch = chr(int(byte_bits, 2))
                    if ch.isprintable() or ch in '\n\r\t':
                        decoded += ch
            if decoded.strip():
                print(f"\n    Possible decoded message:")
                print(f"    {decoded[:500]}")
    else:
        print("No zero-width Unicode steganography detected.")
except Exception as e:
    print(f"Error: {e}")
`, fp)
				return exec.Command("python3", "-c", script)
			},
		},
		{
			Name:           "spammimic",
			Binary:         "python3",
			Category:       "text",
			SupportedTypes: allTypes,
			BuildCmd: func(fp string, opts RunOpts) *exec.Cmd {
				// Detect spammimic-style encoded text
				script := fmt.Sprintf(`
import re
try:
    with open('%s', 'r', errors='ignore') as f:
        data = f.read()
    spam_indicators = [
        'dear friend', 'make money', 'limited time offer', 'click here',
        'act now', 'free', 'winner', 'congratulations', 'urgent',
        'dear sir', 'opportunity', 'investment', 'discount',
        'earn extra', 'no obligation', 'risk free', 'special promotion',
        'be your own boss', 'work from home', 'double your',
    ]
    score = 0
    matches = []
    lower = data.lower()
    for phrase in spam_indicators:
        count = lower.count(phrase)
        if count > 0:
            score += count
            matches.append(f"    '{phrase}': {count}x")
    if score >= 3:
        print("[!] SpamMimic-style steganography SUSPECTED!")
        print(f"    Spam score: {score} (threshold: 3)")
        print(f"    Matching phrases:")
        for m in matches[:15]:
            print(m)
        print(f"\n    Try decoding at: https://www.spammimic.com/decode.shtml")
    else:
        print("No SpamMimic steganography patterns detected.")
except Exception as e:
    print(f"Error: {e}")
`, fp)
				return exec.Command("python3", "-c", script)
			},
		},
	}
}
