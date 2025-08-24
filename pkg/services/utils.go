package services

import (
	"regexp"
	"strconv"
	"strings"
)

var patterns = []*regexp.Regexp{
	// S01E02, s01.e02, s01-e02
	regexp.MustCompile(`(?i)s(\d{1,2})[.\s_-]*e(\d{1,2})`),
	// 1x02 style
	regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,2})`),
	// Verbose: Season 1 Episode 2
	regexp.MustCompile(`(?i)season[ ._-]?(\d{1,2}).*episode[ ._-]?(\d{1,2})`),
	// Season only
	regexp.MustCompile(`(?i)s(\d{1,2})(?:[^0-9]|$)`),
	regexp.MustCompile(`(?i)season[ ._-]?(\d{1,2})`),
	// Episode only
	regexp.MustCompile(`(?i)e(\d{1,2})(?:[^0-9]|$)`),
}

func ExtractSeasonEpisode(filename string) (season, episode int) {
	lower := strings.ToLower(filename)

	for _, re := range patterns {
		m := re.FindStringSubmatch(lower)
		if len(m) == 0 {
			continue
		}

		switch len(m) {
		case 3: // season + episode
			season, _ = strconv.Atoi(m[1])
			episode, _ = strconv.Atoi(m[2])
			return season, episode
		case 2: // season-only or episode-only
			num, _ := strconv.Atoi(m[1])
			if strings.Contains(re.String(), "e(") { // episode-only regex
				return 0, num
			}
			return num, 0
		}
	}
	return 0, 0
}
