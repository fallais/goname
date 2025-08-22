package database

import "goname/internal/models"

type VideoDatabase interface {
	SearchMovie(title string, year int) (*models.Movie, error)
	SearchTVShow(title string, year int) (*models.TVShow, error)
	GetEpisode(showID, season, episode int) (*models.Episode, error)
}
