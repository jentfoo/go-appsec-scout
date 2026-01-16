package sources

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRapidDNS(t *testing.T) {
	t.Parallel()

	t.Run("registered", func(t *testing.T) {
		src := ByName("rapiddns")
		require.NotNil(t, src)
		assert.Equal(t, Subdomain, src.Yields)
	})

	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test")
		}

		ctx := t.Context()
		client := &http.Client{Timeout: 60 * time.Second}
		subdomains, _, errors := collectResults(RapidDNS.Run(ctx, client, "github.com", ""))

		if len(errors) > 0 {
			t.Logf("errors: %v", errors)
		}

		t.Logf("found %d subdomains", len(subdomains))
		assert.NotEmpty(t, subdomains)
		assertResults(t, subdomains, "rapiddns", Subdomain)
	})
}

func TestRapidDNSPagePattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name:     "single_page_link",
			html:     `<a class="page-link" href="/subdomain/example.com?page=2">`,
			expected: []string{"2"},
		},
		{
			name:     "multiple_page_links",
			html:     `<a class="page-link" href="/subdomain/example.com?page=1"><a class="page-link" href="/subdomain/example.com?page=5">`,
			expected: []string{"1", "5"},
		},
		{
			name:     "no_page_links",
			html:     `<a href="/other">No pagination</a>`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := rapidDNSPagePattern.FindAllStringSubmatch(tt.html, -1)
			var pages []string
			for _, m := range matches {
				if len(m) >= 2 {
					pages = append(pages, m[1])
				}
			}
			assert.Equal(t, tt.expected, pages)
		})
	}
}
