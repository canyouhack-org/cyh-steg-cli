package detector

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileCategory represents the type of file being analyzed
type FileCategory string

const (
	CategoryPNG     FileCategory = "png"
	CategoryJPG     FileCategory = "jpg"
	CategoryBMP     FileCategory = "bmp"
	CategoryGIF     FileCategory = "gif"
	CategoryTIFF    FileCategory = "tiff"
	CategoryWEBP    FileCategory = "webp"
	CategoryWAV     FileCategory = "wav"
	CategoryMP3     FileCategory = "mp3"
	CategoryFLAC    FileCategory = "flac"
	CategoryOGG     FileCategory = "ogg"
	CategoryAU      FileCategory = "au"
	CategoryUnknown FileCategory = "unknown"
)

// FileInfo contains detected info about the target file
type FileInfo struct {
	Path      string
	Name      string
	Size      int64
	Category  FileCategory
	MimeType  string
	Extension string
}

// IsImage returns true if the file is an image
func (f *FileInfo) IsImage() bool {
	switch f.Category {
	case CategoryPNG, CategoryJPG, CategoryBMP, CategoryGIF, CategoryTIFF, CategoryWEBP:
		return true
	}
	return false
}

// IsAudio returns true if the file is an audio file
func (f *FileInfo) IsAudio() bool {
	switch f.Category {
	case CategoryWAV, CategoryMP3, CategoryFLAC, CategoryOGG, CategoryAU:
		return true
	}
	return false
}

// Detect analyzes a file and returns its FileInfo
func Detect(filePath string) (*FileInfo, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	f, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	// Read first 512 bytes for MIME detection
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}
	buf = buf[:n]

	mimeType := http.DetectContentType(buf)
	ext := strings.ToLower(filepath.Ext(absPath))

	category := detectCategory(mimeType, ext, buf)

	return &FileInfo{
		Path:      absPath,
		Name:      stat.Name(),
		Size:      stat.Size(),
		Category:  category,
		MimeType:  mimeType,
		Extension: ext,
	}, nil
}

func detectCategory(mime, ext string, header []byte) FileCategory {
	// Check magic bytes first for accuracy
	if len(header) >= 8 {
		// PNG: 89 50 4E 47
		if header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 {
			return CategoryPNG
		}
		// JPEG: FF D8 FF
		if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
			return CategoryJPG
		}
		// BMP: 42 4D
		if header[0] == 0x42 && header[1] == 0x4D {
			return CategoryBMP
		}
		// GIF: 47 49 46 38
		if header[0] == 0x47 && header[1] == 0x49 && header[2] == 0x46 && header[3] == 0x38 {
			return CategoryGIF
		}
		// TIFF: 49 49 2A 00 or 4D 4D 00 2A
		if (header[0] == 0x49 && header[1] == 0x49 && header[2] == 0x2A && header[3] == 0x00) ||
			(header[0] == 0x4D && header[1] == 0x4D && header[2] == 0x00 && header[3] == 0x2A) {
			return CategoryTIFF
		}
		// RIFF....WAVE (WAV)
		if header[0] == 0x52 && header[1] == 0x49 && header[2] == 0x46 && header[3] == 0x46 &&
			len(header) >= 12 && header[8] == 0x57 && header[9] == 0x41 && header[10] == 0x56 && header[11] == 0x45 {
			return CategoryWAV
		}
		// RIFF....WEBP
		if header[0] == 0x52 && header[1] == 0x49 && header[2] == 0x46 && header[3] == 0x46 &&
			len(header) >= 12 && header[8] == 0x57 && header[9] == 0x45 && header[10] == 0x42 && header[11] == 0x50 {
			return CategoryWEBP
		}
		// FLAC: 66 4C 61 43
		if header[0] == 0x66 && header[1] == 0x4C && header[2] == 0x61 && header[3] == 0x43 {
			return CategoryFLAC
		}
		// OGG: 4F 67 67 53
		if header[0] == 0x4F && header[1] == 0x67 && header[2] == 0x67 && header[3] == 0x53 {
			return CategoryOGG
		}
		// MP3: FF FB or FF F3 or FF F2 or ID3
		if (header[0] == 0xFF && (header[1]&0xE0) == 0xE0) ||
			(header[0] == 0x49 && header[1] == 0x44 && header[2] == 0x33) {
			return CategoryMP3
		}
		// AU: 2E 73 6E 64
		if header[0] == 0x2E && header[1] == 0x73 && header[2] == 0x6E && header[3] == 0x64 {
			return CategoryAU
		}
	}

	// Fallback to extension-based detection
	switch ext {
	case ".png":
		return CategoryPNG
	case ".jpg", ".jpeg", ".jfif", ".jpe":
		return CategoryJPG
	case ".bmp":
		return CategoryBMP
	case ".gif":
		return CategoryGIF
	case ".tiff", ".tif":
		return CategoryTIFF
	case ".webp":
		return CategoryWEBP
	case ".wav":
		return CategoryWAV
	case ".mp3":
		return CategoryMP3
	case ".flac":
		return CategoryFLAC
	case ".ogg":
		return CategoryOGG
	case ".au":
		return CategoryAU
	}

	return CategoryUnknown
}
