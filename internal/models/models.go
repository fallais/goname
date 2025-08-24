package models

import "time"

// VideoFile represents a video file to be renamed
type VideoFile struct {
	Path         string
	OriginalName string
	NewName      string
	Size         int64
	ModTime      time.Time
	MediaType    MediaType
}

// MediaType represents the type of media (movie, tv show, etc.)
type MediaType string

const (
	MediaTypeMovie  MediaType = "movie"
	MediaTypeTVShow MediaType = "tv"
	MediaTypeAnime  MediaType = "anime"
)

// MovieInfo represents movie information from TMDB
type MovieInfo struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title"`
	ReleaseDate   string `json:"release_date"`
	Year          int    `json:"-"`
	Overview      string `json:"overview"`
	PosterPath    string `json:"poster_path"`
	IMDBId        string `json:"imdb_id"`
}

// TVShowInfo represents TV show information
type TVShowInfo struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	FirstAirDate string `json:"first_air_date"`
	Year         int    `json:"-"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
}

// EpisodeInfo represents episode information
type EpisodeInfo struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	SeasonNumber  int    `json:"season_number"`
	EpisodeNumber int    `json:"episode_number"`
	AirDate       string `json:"air_date"`
	StillPath     string `json:"still_path"`
}

// RenameResult represents the result of a rename operation
type RenameResult struct {
	VideoFile   VideoFile
	Success     bool
	Error       error
	MediaInfo   interface{} // Can be MovieInfo, TVShowInfo, etc.
	NewFileName string
}

// StateEntry represents a single rename operation in the state
type StateEntry struct {
	ID           string      `json:"id"`
	Timestamp    time.Time   `json:"timestamp"`
	OriginalPath string      `json:"original_path"`
	NewPath      string      `json:"new_path"`
	OriginalName string      `json:"original_name"`
	NewName      string      `json:"new_name"`
	MediaInfo    interface{} `json:"media_info,omitempty"`
	Reverted     bool        `json:"reverted"`
}

// State represents the complete state file
type State struct {
	Version string       `json:"version"`
	Entries []StateEntry `json:"entries"`
}
