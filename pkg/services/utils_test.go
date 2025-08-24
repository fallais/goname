package services

import (
	"testing"
)

func TestExtractSeasonEpisode(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		expectedSeason  int
		expectedEpisode int
	}{
		{
			name:            "Standard S01E01 format",
			filename:        "show.s01e01.720p.mkv",
			expectedSeason:  1,
			expectedEpisode: 1,
		},
		{
			name:            "Standard S01E01 format uppercase",
			filename:        "SHOW.S01E01.720P.MKV",
			expectedSeason:  1,
			expectedEpisode: 1,
		},
		{
			name:            "Mixed case S01E01 format",
			filename:        "Show.S01e01.720p.mkv",
			expectedSeason:  1,
			expectedEpisode: 1,
		},
		{
			name:            "Double digit season and episode",
			filename:        "show.s12e34.1080p.mkv",
			expectedSeason:  12,
			expectedEpisode: 34,
		},
		{
			name:            "Single digit season, double digit episode",
			filename:        "show.s01e15.mkv",
			expectedSeason:  1,
			expectedEpisode: 15,
		},
		{
			name:            "Double digit season, single digit episode",
			filename:        "show.s15e01.mkv",
			expectedSeason:  15,
			expectedEpisode: 1,
		},
		{
			name:            "Season and episode in middle of filename",
			filename:        "my.favorite.show.s02e08.some.quality.mkv",
			expectedSeason:  2,
			expectedEpisode: 8,
		},
		{
			name:            "Season and episode at start of filename",
			filename:        "s03e12.show.name.mkv",
			expectedSeason:  3,
			expectedEpisode: 12,
		},
		{
			name:            "Season and episode with dots",
			filename:        "show.name.s04.e05.mkv",
			expectedSeason:  4,
			expectedEpisode: 5,
		},
		{
			name:            "No season/episode pattern",
			filename:        "movie.2023.1080p.mkv",
			expectedSeason:  0,
			expectedEpisode: 0,
		},
		{
			name:            "Incomplete pattern - only season",
			filename:        "show.s01.mkv",
			expectedSeason:  1,
			expectedEpisode: 0,
		},
		{
			name:            "Edge case - s at end of filename",
			filename:        "shows",
			expectedSeason:  0,
			expectedEpisode: 0,
		},
		{
			name:            "Edge case - empty filename",
			filename:        "",
			expectedSeason:  0,
			expectedEpisode: 0,
		},
		{
			name:            "Season 0 Episode 0",
			filename:        "show.s00e00.mkv",
			expectedSeason:  0,
			expectedEpisode: 0,
		},
		{
			name:            "Season with leading zero",
			filename:        "show.s09e12.mkv",
			expectedSeason:  9,
			expectedEpisode: 12,
		},
		{
			name:            "Episode with leading zero",
			filename:        "show.s12e09.mkv",
			expectedSeason:  12,
			expectedEpisode: 9,
		},
		{
			name:            "Multiple s patterns - should take first valid one",
			filename:        "show.s01e02.some.s03e04.mkv",
			expectedSeason:  1,
			expectedEpisode: 2,
		},
		{
			name:            "Very short filename with pattern",
			filename:        "s1e2",
			expectedSeason:  1,
			expectedEpisode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			season, episode := ExtractSeasonEpisode(tt.filename)
			if season != tt.expectedSeason {
				t.Errorf("extractSeasonEpisode(%q) season = %d, want %d", tt.filename, season, tt.expectedSeason)
			}
			if episode != tt.expectedEpisode {
				t.Errorf("extractSeasonEpisode(%q) episode = %d, want %d", tt.filename, episode, tt.expectedEpisode)
			}
		})
	}
}

// TestExtractSeasonEpisodeBenchmark provides a benchmark for the function
func BenchmarkExtractSeasonEpisode(b *testing.B) {
	testCases := []string{
		"show.s01e01.720p.mkv",
		"my.favorite.show.s12e34.1080p.mkv",
		"movie.2023.1080p.mkv",
		"show.name.s04.e05.mkv",
	}

	for i := 0; i < b.N; i++ {
		for _, filename := range testCases {
			ExtractSeasonEpisode(filename)
		}
	}
}
