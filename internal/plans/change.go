package plans

type Action rune

const (
	ActionNoop   Action = 0
	ActionRename Action = '~'
)

type FileInfo struct {
	Path     string `json:"path"`
	FileName string `json:"file_name"`
}

// Change represents a proposed change to a video file.
type Change struct {
	ID string `json:"id"`

	// Action to be performed
	Action Action `json:"action"`

	Before FileInfo `json:"before"`
	After  FileInfo `json:"after"`
}
