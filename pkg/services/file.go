package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goname/internal/models"
)

// FileService handles file operations and scanning
type FileService struct {
	supportedExtensions map[string]bool
}

// NewFileService creates a new file service instance
func NewFileService() *FileService {
	// Common video file extensions
	extensions := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".m4v":  true,
		".mpg":  true,
		".mpeg": true,
		".3gp":  true,
		".ogv":  true,
	}

	return &FileService{
		supportedExtensions: extensions,
	}
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
		if !fs.supportedExtensions[ext] {
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

// GenerateMovieFileName generates a standardized filename for a movie
func (fs *FileService) GenerateMovieFileName(movie *models.Movie, originalPath string) string {
	ext := filepath.Ext(originalPath)

	// Format: "Movie Title (Year).ext"
	var filename string
	if movie.ReleaseDate.Year() > 0 {
		filename = sanitizeFilename(movie.Title) + " (" + string(rune(movie.ReleaseDate.Year())) + ")" + ext
	} else {
		filename = sanitizeFilename(movie.Title) + ext
	}

	return filename
}

// GenerateTVShowFileName generates a standardized filename for a TV show episode
func (fs *FileService) GenerateTVShowFileName(show *models.TVShow, episode *models.Episode, originalPath string) string {
	ext := filepath.Ext(originalPath)

	// Format: "Show Name - S01E01 - Episode Name.ext"
	filename := sanitizeFilename(show.Name) +
		" - S" + padNumber(episode.Season, 2) +
		"E" + padNumber(episode.Episode, 2)

	if episode.Title != "" {
		filename += " - " + sanitizeFilename(episode.Title)
	}

	filename += ext
	return filename
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
