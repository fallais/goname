package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"goname/internal/models"

	tmdb "github.com/cyruzin/golang-tmdb"
)

// TMDBService handles interactions with The Movie Database API
type TMDBService struct {
	client *tmdb.Client
	apiKey string
}

// NewTMDBService creates a new TMDB service instance
func NewTMDBService(apiKey string) (*TMDBService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("TMDB API key is required")
	}

	client, err := tmdb.InitV4(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize TMDB client: %w", err)
	}

	return &TMDBService{
		client: client,
		apiKey: apiKey,
	}, nil
}

// SearchMovie searches for movies by title with improved matching
func (s *TMDBService) SearchMovie(title string, year int) (*models.Movie, error) {
	// First try the exact title
	movie, err := s.searchMovieWithQuery(title, year)
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
				if movie, err := s.searchMovieWithQuery(query, year); err == nil {
					return movie, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no movies found for title: %s (tried various word combinations)", title)
}

// searchMovieWithQuery performs a single search query to TMDB
func (s *TMDBService) searchMovieWithQuery(title string, year int) (*models.Movie, error) {
	options := map[string]string{}

	if year > 0 {
		options["year"] = strconv.Itoa(year)
	}

	results, err := s.client.GetSearchMovies(title, options)
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

// SearchTVShow searches for TV shows by name with improved matching
func (s *TMDBService) SearchTVShow(name string, year int) (*models.TVShow, error) {
	// First try the exact name
	show, err := s.searchTVShowWithQuery(name, year)
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
				if show, err := s.searchTVShowWithQuery(query, year); err == nil {
					return show, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no TV shows found for name: %s (tried various word combinations)", name)
}

// searchTVShowWithQuery performs a single search query to TMDB
func (s *TMDBService) searchTVShowWithQuery(name string, year int) (*models.TVShow, error) {
	options := map[string]string{}

	if year > 0 {
		options["first_air_date_year"] = strconv.Itoa(year)
	}

	results, err := s.client.GetSearchTVShow(name, options)
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

// GetEpisode gets episode information for a specific TV show
func (s *TMDBService) GetEpisode(tvShowID, seasonNumber, episodeNumber int) (*models.Episode, error) {
	episode, err := s.client.GetTVEpisodeDetails(tvShowID, seasonNumber, episodeNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode details: %w", err)
	}

	var airDate time.Time
	if episode.AirDate != "" {
		if date, err := time.Parse("2006-01-02", episode.AirDate); err == nil {
			airDate = date
		}
	}

	return &models.Episode{
		Title:   episode.Name,
		Season:  episode.SeasonNumber,
		Episode: episode.EpisodeNumber,
		AirDate: airDate,
		//StillPath: episode.StillPath,
		Database: models.Database{
			IMDBId: strconv.FormatInt(episode.ID, 10),
		},
	}, nil
}

// ExtractYear tries to extract a year from a filename
func ExtractYear(filename string) int {
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
}
