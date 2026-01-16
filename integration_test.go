package scout

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-harden/scout/sources"
)

func TestIntegrationQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	const domain = "github.com"
	ctx := t.Context()

	results, err := Collect(Query(ctx, domain))

	assert.NotEmpty(t, results, "complete failure: %v", err)

	// Validate deduplication by checking for unique values (case-insensitive)
	seen := make(map[string]sources.Result)
	for _, r := range results {
		normalized := strings.ToLower(strings.TrimSpace(r.Value))
		if existing, found := seen[normalized]; found {
			t.Errorf("found duplicate result: %q from source %q (previously seen from source %q)",
				r.Value, r.Source, existing.Source)
		}
		seen[normalized] = r
	}

	// All results should have a source name
	for _, r := range results {
		assert.NotEmpty(t, r.Source)
	}

	// All results should have a valid type
	for _, r := range results {
		assert.True(t, r.Type == sources.Subdomain || r.Type == sources.URL)
	}

	// All results should have a non-empty value
	for _, r := range results {
		assert.NotEmpty(t, r.Value)
	}

	// Subdomains should contain the target domain
	for _, r := range results {
		if r.Type == sources.Subdomain {
			assert.Contains(t, r.Value, domain)
		}
	}

	// URLs should start with http:// or https://
	for _, r := range results {
		if r.Type == sources.URL {
			hasHTTP := strings.HasPrefix(r.Value, "http://")
			hasHTTPS := strings.HasPrefix(r.Value, "https://")
			assert.True(t, hasHTTP || hasHTTPS)
		}
	}
}
