package sources

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiteDossier(t *testing.T) {
	t.Parallel()

	t.Run("registered", func(t *testing.T) {
		src := ByName("sitedossier")
		require.NotNil(t, src)
		assert.Equal(t, Subdomain, src.Yields)
	})

	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test")
		}

		ctx := t.Context()
		client := &http.Client{Timeout: 60 * time.Second}
		subdomains, _, errors := collectResults(SiteDossier.Run(ctx, client, "github.com", ""))

		if len(errors) > 0 {
			t.Logf("errors: %v", errors)
		}

		// SiteDossier may be unreliable
		t.Logf("found %d subdomains", len(subdomains))
		assertResults(t, subdomains, "sitedossier", Subdomain)
	})
}

func TestSiteDossierNextPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "has_next_link",
			html:     `<a href="/parentdomain/example.com/2"><b>Next`,
			expected: "/parentdomain/example.com/2",
		},
		{
			name:     "no_next_link",
			html:     `<div>No more pages</div>`,
			expected: "",
		},
		{
			name:     "different_link",
			html:     `<a href="/other"><span>Other</span></a>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if matches := siteDossierNextPattern.FindStringSubmatch(tt.html); len(matches) >= 2 {
				result = matches[1]
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
