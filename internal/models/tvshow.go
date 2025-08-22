package models

import "time"

type TVShow struct {
	Name         string    `json:"name"`
	OriginalName string    `json:"original_name"`
	FirstAirDate time.Time `json:"first_air_date"`
	Genre        string    `json:"genre"`
	Director     string    `json:"director"`
	Seasons      int       `json:"seasons"`
	Episodes     int       `json:"episodes"`
	Database     Database  `json:"database"`
}
