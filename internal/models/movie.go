package models

import "time"

type Movie struct {
	Title         string    `json:"title"`
	OriginalTitle string    `json:"original_title"`
	ReleaseDate   time.Time `json:"release_date"`
	Genre         Genre     `json:"genre"`
	Director      string    `json:"director"`
	Database      Database  `json:"database"`
}

type Database struct {
	IMDBId string `json:"imdb_id"`
}

type Confidence struct {
	Score float64 `json:"score"`
}
