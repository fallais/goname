package services

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"goname/internal/models"
)

var SupportedExtensions = []string{
	".mp4",
	".mkv",
	".avi",
	".mov",
	".wmv",
	".flv",
	".webm",
	".m4v",
	".mpg",
	".mpeg",
	".3gp",
	".ogv",
}

// FileService handles file operations and scanning
type FileService struct {
	supportedExtensions []string
	tvShowTemplate      string
	movieTemplate       string
}

// NewFileService creates a new file service instance
func NewFileService() *FileService {
	return &FileService{
		supportedExtensions: SupportedExtensions,
		tvShowTemplate:      TVShowTemplateDefault,
		movieTemplate:       MovieTemplateDefault,
	}
}

// NewFileServiceWithTemplates creates a new file service instance with custom templates
func NewFileServiceWithTemplates(tvTemplate, movieTemplate string) *FileService {
	return &FileService{
		supportedExtensions: SupportedExtensions,
		tvShowTemplate:      tvTemplate,
		movieTemplate:       movieTemplate,
	}
}

// SetTVShowTemplate sets the template for TV show filename generation
func (fs *FileService) SetTVShowTemplate(template string) error {
	if err := fs.ValidateTemplate(template); err != nil {
		return err
	}
	fs.tvShowTemplate = template
	return nil
}

// SetMovieTemplate sets the template for movie filename generation
func (fs *FileService) SetMovieTemplate(template string) error {
	if err := fs.ValidateTemplate(template); err != nil {
		return err
	}
	fs.movieTemplate = template
	return nil
}

// GetTVShowTemplate returns the current TV show template
func (fs *FileService) GetTVShowTemplate() string {
	return fs.tvShowTemplate
}

// GetMovieTemplate returns the current movie template
func (fs *FileService) GetMovieTemplate() string {
	return fs.movieTemplate
}

// ScanDirectory scans a directory for video files
func (fs *FileService) ScanDirectory(dirPath string, recursive bool) ([]models.VideoFile, error) {
	var videoFiles []models.VideoFile

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// If not recursive and this is a subdirectory, skip it
			if !recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has a supported video extension
		ext := strings.ToLower(filepath.Ext(path))
		if !slices.Contains(fs.supportedExtensions, ext) {
			return nil
		}

		videoFile := models.VideoFile{
			Path:         path,
			OriginalName: info.Name(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
		}

		// Try to determine media type from filename
		videoFile.MediaType = fs.guessMediaType(info.Name())

		videoFiles = append(videoFiles, videoFile)
		return nil
	})

	return videoFiles, err
}

// guessMediaType tries to guess if a file is a movie or TV show based on filename patterns
func (fs *FileService) guessMediaType(filename string) models.MediaType {
	filename = strings.ToLower(filename)

	// Common TV show patterns
	tvPatterns := []string{
		"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", // Season patterns
		"e0", "e1", "e2", "e3", "e4", "e5", "e6", "e7", "e8", "e9", // Episode patterns
		"episode", "ep", "season",
		"x0", "x1", "x2", "x3", "x4", "x5", "x6", "x7", "x8", "x9", // SxxExx pattern
	}

	for _, pattern := range tvPatterns {
		if strings.Contains(filename, pattern) {
			return models.MediaTypeTVShow
		}
	}

	// Default to movie if no TV patterns found
	return models.MediaTypeMovie
}

// RenameFile renames a file from old path to new path
func (fs *FileService) RenameFile(oldPath, newPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.Rename(oldPath, newPath)
}

// sanitizeFilename removes or replaces characters that are not allowed in filenames
func sanitizeFilename(filename string) string {
	// Replace invalid characters with safe alternatives
	replacements := map[string]string{
		":":  " -",
		"*":  "",
		"?":  "",
		"\"": "'",
		"<":  "",
		">":  "",
		"|":  "",
		"/":  "",
		"\\": "",
	}

	for old, new := range replacements {
		filename = strings.ReplaceAll(filename, old, new)
	}

	// Remove multiple spaces and trim
	filename = strings.Join(strings.Fields(filename), " ")
	return strings.TrimSpace(filename)
}

// padNumber pads a number with leading zeros
func padNumber(num, width int) string {
	return fmt.Sprintf("%0*d", width, num)
}
