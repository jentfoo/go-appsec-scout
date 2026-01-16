package sources

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommonCrawl(t *testing.T) {
	t.Parallel()

	t.Run("registered", func(t *testing.T) {
		src := ByName("commoncrawl")
		require.NotNil(t, src)
		assert.Equal(t, Subdomain|URL, src.Yields)
	})

	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test")
		}

		ctx := t.Context()
		client := &http.Client{Timeout: 120 * time.Second}
		subdomains, urls, errors := collectResults(CommonCrawl.Run(ctx, client, "github.com", ""))

		if len(errors) > 0 {
			t.Logf("errors: %v", errors)
		}

		t.Logf("found %d subdomains, %d urls", len(subdomains), len(urls))
		assert.NotEmpty(t, subdomains)
		assert.NotEmpty(t, urls)
		assertResults(t, subdomains, "commoncrawl", Subdomain)
		assertResults(t, urls, "commoncrawl", URL)
	})
}

func TestParseCommonCrawlYear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       string
		expected int
	}{
		{
			name:     "valid_2024",
			id:       "CC-MAIN-2024-10",
			expected: 2024,
		},
		{
			name:     "valid_2023",
			id:       "CC-MAIN-2023-50",
			expected: 2023,
		},
		{
			name:     "invalid_format",
			id:       "invalid",
			expected: 0,
		},
		{
			name:     "missing_year",
			id:       "CC-MAIN",
			expected: 0,
		},
		{
			name:     "non_numeric_year",
			id:       "CC-MAIN-abc-10",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseCommonCrawlYear(tt.id))
		})
	}
}

func TestDecodeCommonCrawlURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain_url",
			input:    "https://example.com/path",
			expected: "https://example.com/path",
		},
		{
			name:     "encoded_slash_lowercase",
			input:    "https://example.com/path%2ffile",
			expected: "https://example.com/path/file",
		},
		{
			name:     "encoded_slash_uppercase",
			input:    "https://example.com/path%2Ffile",
			expected: "https://example.com/path/file",
		},
		{
			name:     "artifact_25_percent",
			input:    "https://example.com/%25query",
			expected: "https://example.com/%query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, decodeCommonCrawlURL(tt.input))
		})
	}
}
