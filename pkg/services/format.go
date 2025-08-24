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
	TVShowTemplateDefault = PlexFormatTVShow

	// Movie templates
	MovieTemplateDefault = PlexFormatMovie

	// PlexFormatTVShow is : ShowName - S01E01 - First Episode
	PlexFormatTVShow = "{{.Name}} - S{{printf \"%02d\" .Season}}E{{printf \"%02d\" .Episode}} - {{.Title}}"

	// PlexFormatMovie is : MovieName (2001)
	PlexFormatMovie = "{{.Name}} ({{.Year}})"

	// KodiFormatTVShow is : ShowName (2001) - 1x01 - First Episode
	KodiFormatTVShow = "{{.Name}} ({{.Year}}) - {{.Title}} {{.Season}}x{{.Episode}})"

	// EmbyFormatTVShow is : ShowName (2001) - S01E01 - First Episode
	EmbyFormatTVShow = "{{.Name}} ({{.Year}}) - S{{printf \"%02d\" .Season}}E{{printf \"%02d\" .Episode}} - {{.Title}}"
)

// TemplateData represents the data available for filename templates
type TemplateData struct {
	// Common fields
	Name  string // Movie title or TV show name
	Title string // Episode title for TV shows, empty for movies
	Year  int    // Release year

	// TV Show specific fields
	Season  int // Season number (0 for movies)
	Episode int // Episode number (0 for movies)

	// Movie specific fields
	Director string // Director name
	Genre    string // Genre
}

// GenerateFileNameFromTemplate generates a filename using the provided template and data
// The extension is automatically appended and not part of the template
func (fs *FileService) GenerateFileNameFromTemplate(templateStr string, data *TemplateData, extension string) string {
	tmpl, err := template.New("filename").Parse(templateStr)
	if err != nil {
		// TODO return error
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		// TODO return error
	}

	// Append the extension after template processing
	filename := sanitizeFilename(buf.String()) + extension
	return filename
}

// GenerateMovieFileNameWithTemplate generates a movie filename using a custom template
func (fs *FileService) GenerateMovieFileNameWithTemplate(movie *models.Movie, originalPath, templateStr string) string {
	return fs.GenerateFileNameFromTemplate(templateStr, &TemplateData{
		Name: movie.Title,
		Year: movie.ReleaseDate.Year(),
		// TV show fields are 0 for movies
		Season:  0,
		Episode: 0,
		// Additional movie fields
		Director: movie.Director,
		Genre:    string(movie.Genre),
	}, filepath.Ext(originalPath))
}

// GenerateTVShowFileNameWithTemplate generates a TV show episode filename using a custom template
func (fs *FileService) GenerateTVShowFileNameWithTemplate(show *models.TVShow, episode *models.Episode, originalPath, templateStr string) string {
	return fs.GenerateFileNameFromTemplate(templateStr, &TemplateData{
		Name:    show.Name,
		Title:   episode.Title,
		Season:  episode.Season,
		Episode: episode.Episode,
		Year:    show.FirstAirDate.Year(),
		// Additional fields
		Director: "", // Not available for episodes
		Genre:    "", // Could be added if available
	}, filepath.Ext(originalPath))
}

// GenerateMovieFileName generates a standardized filename for a movie using template
func (fs *FileService) GenerateMovieFileName(movie *models.Movie, originalPath string) string {
	return fs.GenerateFileNameFromTemplate(fs.movieTemplate, &TemplateData{
		Name: movie.Title,
		Year: movie.ReleaseDate.Year(),
		// TV show fields are 0 for movies
		Season:  0,
		Episode: 0,
		// Additional movie fields
		Director: movie.Director,
		Genre:    string(movie.Genre),
	}, filepath.Ext(originalPath))
}

// GenerateTVShowFileName generates a standardized filename for a TV show episode using template
func (fs *FileService) GenerateTVShowFileName(show *models.TVShow, episode *models.Episode, originalPath string) string {
	return fs.GenerateFileNameFromTemplate(fs.tvShowTemplate, &TemplateData{
		Name:    show.Name,
		Title:   episode.Title,
		Season:  episode.Season,
		Episode: episode.Episode,
		Year:    show.FirstAirDate.Year(),
		// Additional fields
		Director: "", // Not available for episodes
		Genre:    "", // Could be added if available
	}, filepath.Ext(originalPath))
}
