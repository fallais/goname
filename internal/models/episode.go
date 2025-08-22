package models

import "time"

// Episode represents a TV show episode
type Episode struct {
	Title     string
	Season    int
	Episode   int
	AirDate   time.Time
	Summary   string
	Thumbnail string
	Database  Database
}
