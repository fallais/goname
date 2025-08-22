package tmdb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"goname/internal/models"
	"goname/pkg/database"

	tmdb "github.com/cyruzin/golang-tmdb"
)

type theMovieDatabase struct {
	client *tmdb.Client
	apiKey string
}

// New creates a new TMDB service instance
func New(apiKey string) (database.VideoDatabase, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("TMDB API key is required")
	}

	client, err := tmdb.InitV4(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize TMDB client: %w", err)
	}

	return theMovieDatabase{
		client: client,
		apiKey: apiKey,
	}, nil
}

// SearchMovie searches for movies by title with improved matching
func (d theMovieDatabase) SearchMovie(title string, year int) (*models.Movie, error) {
	// First try the exact title
	movie, err := d.searchMovieWithQuery(title, year)
	if err == nil {
		return movie, nil
	}

	// If no results, try searching with individual words
	words := strings.Fields(title)
	if len(words) > 1 {
		// Try different combinations of words, starting with longer combinations
		for i := len(words); i >= 1; i-- {
			for j := 0; j <= len(words)-i; j++ {
				query := strings.Join(words[j:j+i], " ")
				if movie, err := d.searchMovieWithQuery(query, year); err == nil {
					return movie, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no movies found for title: %s (tried various word combinations)", title)
}

// SearchTVShow searches for TV shows by name with improved matching
func (d theMovieDatabase) SearchTVShow(name string, year int) (*models.TVShow, error) {
	// First try the exact name
	show, err := d.searchTVShowWithQuery(name, year)
	if err == nil {
		return show, nil
	}

	// If no results, try searching with individual words
	words := strings.Fields(name)
	if len(words) > 1 {
		// Try different combinations of words, starting with longer combinations
		for i := len(words); i >= 1; i-- {
			for j := 0; j <= len(words)-i; j++ {
				query := strings.Join(words[j:j+i], " ")
				if show, err := d.searchTVShowWithQuery(query, year); err == nil {
					return show, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no TV shows found for name: %s (tried various word combinations)", name)
}

// GetEpisode gets episode information for a specific TV show
func (d theMovieDatabase) GetEpisode(tvShowID, seasonNumber, episodeNumber int) (*models.Episode, error) {
	episode, err := d.client.GetTVEpisodeDetails(tvShowID, seasonNumber, episodeNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode details: %w", err)
	}

	// Parse air date
	var airDate time.Time
	if episode.AirDate != "" {
		if date, err := time.Parse("2006-01-02", episode.AirDate); err == nil {
			airDate = date
		}
	}

	return &models.Episode{
		Title:     episode.Name,
		Season:    episode.SeasonNumber,
		Episode:   episode.EpisodeNumber,
		AirDate:   airDate,
		Summary:   episode.Overview,
		Thumbnail: episode.StillPath,
		Database: models.Database{
			IMDBId: strconv.FormatInt(episode.ID, 10),
		},
	}, nil
}

// -------------------- Helper Functions -----------------------------

// searchMovieWithQuery performs a single search query to TMDB
func (d theMovieDatabase) searchMovieWithQuery(title string, year int) (*models.Movie, error) {
	options := map[string]string{}

	if year > 0 {
		options["year"] = strconv.Itoa(year)
	}

	results, err := d.client.GetSearchMovies(title, options)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}

	if len(results.Results) == 0 {
		return nil, fmt.Errorf("no movies found for title: %s", title)
	}

	// Take the first result (most relevant)
	movie := results.Results[0]

	movieInfo := &models.Movie{
		Title:         movie.Title,
		OriginalTitle: movie.OriginalTitle,
		Database: models.Database{
			IMDBId: strconv.FormatInt(movie.ID, 10),
		},
	}

	// Extract year from release date
	if movie.ReleaseDate != "" {
		if date, err := time.Parse("2006-01-02", movie.ReleaseDate); err == nil {
			movieInfo.ReleaseDate = date
		}
	}

	return movieInfo, nil
}

// searchTVShowWithQuery performs a single search query to TMDB
func (d theMovieDatabase) searchTVShowWithQuery(name string, year int) (*models.TVShow, error) {
	options := map[string]string{}

	if year > 0 {
		options["first_air_date_year"] = strconv.Itoa(year)
	}

	results, err := d.client.GetSearchTVShow(name, options)
	if err != nil {
		return nil, fmt.Errorf("failed to search TV shows: %w", err)
	}

	if len(results.Results) == 0 {
		return nil, fmt.Errorf("no TV shows found for name: %s", name)
	}

	// Take the first result (most relevant)
	show := results.Results[0]

	showInfo := &models.TVShow{
		Name:         show.Name,
		OriginalName: show.OriginalName,
		Database: models.Database{
			IMDBId: strconv.FormatInt(show.ID, 10),
		},
	}

	// Extract year from first air date
	if show.FirstAirDate != "" {
		if date, err := time.Parse("2006-01-02", show.FirstAirDate); err == nil {
			showInfo.FirstAirDate = date
		}
	}

	return showInfo, nil
}

// ExtractYear tries to extract a year from a filename
/* func ExtractYear(filename string) int {
	// Look for 4-digit years (1900-2099)
	for _, word := range strings.Fields(filename) {
		if len(word) == 4 {
			if year, err := strconv.Atoi(word); err == nil {
				if year >= 1900 && year <= 2099 {
					return year
				}
			}
		}
	}
	return 0
} */
