package models

// MediaType represents the type of media (movie, tv show, etc.)
type MediaType string

const (
	MediaTypeMovie  MediaType = "movie"
	MediaTypeTVShow MediaType = "tv"
	MediaTypeAnime  MediaType = "anime"
)
