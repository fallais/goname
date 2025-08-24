package database

import (
	"goname/internal/models"
	"strconv"
	"strings"
)

type VideoDatabase interface {
	SearchMovie(title string, year int) (*models.Movie, error)
	SearchTVShow(title string, year int) (*models.TVShow, error)
	GetEpisode(showID, season, episode int) (*models.Episode, error)
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
