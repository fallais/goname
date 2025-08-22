package services

import (
	"regexp"
	"strings"
)

// removeLanguageTags removes language-related tags from filename
func removeLanguageTags(title string) string {
	title = strings.ToLower(title)

	languageTags := []string{
		"multi", "vff", "vfq", "vf", "vo", "vost", "vostfr", "french", "english",
		"german", "spanish", "italian", "portuguese", "russian", "japanese",
		"chinese", "korean", "arabic", "hindi", "truefrench", "fastsub",
	}

	for _, tag := range languageTags {
		title = strings.ReplaceAll(title, tag, " ")
	}

	return title
}

// removeQualityTags removes quality and format-related tags from filename
func removeQualityTags(title string) string {
	title = strings.ToLower(title)

	qualityTags := []string{
		"1080p", "720p", "480p", "4k", "2160p", "hd", "uhd", "fhd",
		"bluray", "webrip", "dvdrip", "bdrip", "hdrip", "tvrip", "web", "dl",
		"x264", "x265", "h264", "h265", "xvid", "divx", "hevc", "avc",
		"dd5.1", "dts", "ac3", "aac", "mp3", "flac", "atmos", "truehd",
		"internal", "proper", "repack", "extended", "unrated", "dc",
		"directors.cut", "final.cut", "theatrical", "imax", "final",
		"amzn", "nf", "hulu", "max", "dsnp", "atvp", "pmtp", "pcok",
	}

	for _, tag := range qualityTags {
		title = strings.ReplaceAll(title, tag, " ")
	}

	return title
}

// removeEpisodeInfo removes season/episode information from filename
func removeEpisodeInfo(title string) string {
	title = strings.ToLower(title)

	// Remove S##E## patterns
	re := regexp.MustCompile(`s\d{1,2}e\d{1,2}`)
	title = re.ReplaceAllString(title, " ")

	// Remove other episode patterns
	episodePatterns := []string{
		"season", "episode", "ep", "part",
	}

	for _, pattern := range episodePatterns {
		title = strings.ReplaceAll(title, pattern, " ")
	}

	return title
}

// removeReleaseGroups removes common release group names and patterns
func removeReleaseGroups(title string) string {
	title = strings.ToLower(title)

	// Common release group patterns (usually at the end after a hyphen)
	releaseGroups := []string{
		"tfa", "yts", "rarbg", "etrg", "sparks", "axxo", "ntsc", "pal",
		"proper", "repack", "internal", "limited", "festival", "readnfo",
		"subbed", "dubbed", "retail", "custom", "uncut", "extended",
		"directors", "theatrical", "imax", "remastered", "criterion",
	}

	for _, group := range releaseGroups {
		title = strings.ReplaceAll(title, group, " ")
	}

	return title
}

// CleanTitle removes common words and characters from titles for better matching
func CleanTitle(title string) string {
	title = strings.ToLower(title)

	// Remove file extensions
	extensions := []string{
		".mkv", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".3gp",
	}
	for _, ext := range extensions {
		title = strings.ReplaceAll(title, ext, "")
	}

	// Remove episode information first (for TV shows)
	title = removeEpisodeInfo(title)

	// Remove language tags
	title = removeLanguageTags(title)

	// Remove quality tags
	title = removeQualityTags(title)

	// Remove release groups
	title = removeReleaseGroups(title)

	// Remove brackets, parentheses and their content
	title = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(title, " ")
	title = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(title, " ")
	title = regexp.MustCompile(`\{.*?\}`).ReplaceAllString(title, " ")

	// Replace separators with spaces
	separators := []string{".", "_", "-", "+", "="}
	for _, sep := range separators {
		title = strings.ReplaceAll(title, sep, " ")
	}

	// Clean up multiple spaces and trim
	words := strings.Fields(title)
	return strings.Join(words, " ")
}
