package services

import (
	"bytes"
	"goname/internal/models"
	"html/template"
	"path/filepath"
)

// Common template formats for different naming conventions
const (
	// TV Show templates
	TVShowTemplateDefault  = "{{.Name}} - S{{printf \"%02d\" .Season}}E{{printf \"%02d\" .Episode}} - {{.Title}}{{.Ext}}"
	TVShowTemplateSimple   = "{{.Name}} S{{printf \"%02d\" .Season}}E{{printf \"%02d\" .Episode}}{{.Ext}}"
	TVShowTemplateWithYear = "{{.Name}} ({{.Year}}) - S{{printf \"%02d\" .Season}}E{{printf \"%02d\" .Episode}} - {{.Title}}{{.Ext}}"

	// Movie templates
	MovieTemplateDefault   = "{{.Name}} ({{.Year}}){{.Ext}}"
	MovieTemplateSimple    = "{{.Name}}{{.Ext}}"
	MovieTemplateWithGenre = "{{.Name}} ({{.Year}}) - {{.Genre}}{{.Ext}}"
)

// TemplateData represents the data available for filename templates
type TemplateData struct {
	// Common fields
	Name  string // Movie title or TV show name
	Title string // Episode title for TV shows, empty for movies
	Year  int    // Release year
	Ext   string // File extension

	// TV Show specific fields
	Season  int // Season number (0 for movies)
	Episode int // Episode number (0 for movies)

	// Movie specific fields
	Director string // Director name
	Genre    string // Genre
}

// GenerateFileNameFromTemplate generates a filename using the provided template and data
func (fs *FileService) GenerateFileNameFromTemplate(templateStr string, data *TemplateData) string {
	tmpl, err := template.New("filename").Parse(templateStr)
	if err != nil {
		// TODO return error
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		// TODO return error
	}

	return sanitizeFilename(buf.String())
}

// ValidateTemplate validates if a template string is valid Go template syntax
func (fs *FileService) ValidateTemplate(templateStr string) error {
	_, err := template.New("test").Parse(templateStr)
	return err
}

// GetAvailableTemplateFields returns a list of available fields for templates
func (fs *FileService) GetAvailableTemplateFields() []string {
	return []string{
		"Name",     // Movie title or TV show name
		"Title",    // Episode title for TV shows, empty for movies
		"Year",     // Release year
		"Ext",      // File extension
		"Season",   // Season number (0 for movies)
		"Episode",  // Episode number (0 for movies)
		"Director", // Director name
		"Genre",    // Genre
	}
}

// GetTemplatePresets returns a map of predefined template presets
func (fs *FileService) GetTemplatePresets() map[string]string {
	return map[string]string{
		"tv_default":       TVShowTemplateDefault,
		"tv_simple":        TVShowTemplateSimple,
		"tv_with_year":     TVShowTemplateWithYear,
		"movie_default":    MovieTemplateDefault,
		"movie_simple":     MovieTemplateSimple,
		"movie_with_genre": MovieTemplateWithGenre,
	}
}

// PreviewTemplateOutput shows what the template would generate with sample data
func (fs *FileService) PreviewTemplateOutput(templateStr string, isMovie bool) (string, error) {
	var data *TemplateData

	if isMovie {
		// Sample movie data
		data = &TemplateData{
			Name:     "Sample Movie",
			Title:    "",
			Year:     2023,
			Ext:      ".mp4",
			Season:   0,
			Episode:  0,
			Director: "Sample Director",
			Genre:    "Action",
		}
	} else {
		// Sample TV show data
		data = &TemplateData{
			Name:     "Sample TV Show",
			Title:    "Sample Episode",
			Year:     2023,
			Ext:      ".mkv",
			Season:   1,
			Episode:  5,
			Director: "",
			Genre:    "",
		}
	}

	if err := fs.ValidateTemplate(templateStr); err != nil {
		return "", err
	}

	return fs.GenerateFileNameFromTemplate(templateStr, data), nil
}

// GenerateMovieFileNameWithTemplate generates a movie filename using a custom template
func (fs *FileService) GenerateMovieFileNameWithTemplate(movie *models.Movie, originalPath, templateStr string) string {
	return fs.GenerateFileNameFromTemplate(templateStr, &TemplateData{
		Name: movie.Title,
		Year: movie.ReleaseDate.Year(),
		Ext:  filepath.Ext(originalPath),
		// TV show fields are 0 for movies
		Season:  0,
		Episode: 0,
		// Additional movie fields
		Director: movie.Director,
		Genre:    string(movie.Genre),
	})
}

// GenerateTVShowFileNameWithTemplate generates a TV show episode filename using a custom template
func (fs *FileService) GenerateTVShowFileNameWithTemplate(show *models.TVShow, episode *models.Episode, originalPath, templateStr string) string {
	return fs.GenerateFileNameFromTemplate(templateStr, &TemplateData{
		Name:    show.Name,
		Title:   episode.Title,
		Season:  episode.Season,
		Episode: episode.Episode,
		Year:    show.FirstAirDate.Year(),
		Ext:     filepath.Ext(originalPath),
		// Additional fields
		Director: "", // Not available for episodes
		Genre:    "", // Could be added if available
	})
}

// GenerateMovieFileName generates a standardized filename for a movie using template
func (fs *FileService) GenerateMovieFileName(movie *models.Movie, originalPath string) string {
	return fs.GenerateFileNameFromTemplate(fs.movieTemplate, &TemplateData{
		Name: movie.Title,
		Year: movie.ReleaseDate.Year(),
		Ext:  filepath.Ext(originalPath),
		// TV show fields are 0 for movies
		Season:  0,
		Episode: 0,
		// Additional movie fields
		Director: movie.Director,
		Genre:    string(movie.Genre),
	})
}

// GenerateTVShowFileName generates a standardized filename for a TV show episode using template
func (fs *FileService) GenerateTVShowFileName(show *models.TVShow, episode *models.Episode, originalPath string) string {
	return fs.GenerateFileNameFromTemplate(fs.tvShowTemplate, &TemplateData{
		Name:    show.Name,
		Title:   episode.Title,
		Season:  episode.Season,
		Episode: episode.Episode,
		Year:    show.FirstAirDate.Year(),
		Ext:     filepath.Ext(originalPath),
		// Additional fields
		Director: "", // Not available for episodes
		Genre:    "", // Could be added if available
	})
}
