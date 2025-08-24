package models

import (
	"time"
)

// VideoFile represents a video file to be renamed
type VideoFile struct {
	Path         string
	OriginalName string
	NewName      string
	Size         int64
	ModTime      time.Time
	MediaType    MediaType
}
