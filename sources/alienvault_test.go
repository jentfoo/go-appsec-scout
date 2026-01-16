package sources

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlienVault(t *testing.T) {
	t.Parallel()

	t.Run("registered", func(t *testing.T) {
		src := ByName("alienvault")
		require.NotNil(t, src)
		assert.Equal(t, Subdomain|URL, src.Yields)
	})

	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test")
		}

		ctx := t.Context()
		client := &http.Client{Timeout: 60 * time.Second}
		subdomains, urls, errors := collectResults(AlienVault.Run(ctx, client, "github.com", ""))

		if len(errors) > 0 {
			t.Logf("errors: %v", errors)
			if strings.Contains(errors[0].Error(), "429") {
				t.Skip("Service overloaded")
			}
		}

		t.Logf("found %d subdomains, %d urls", len(subdomains), len(urls))
		assert.NotEmpty(t, subdomains)
		assert.NotEmpty(t, urls)
		assertResults(t, subdomains, "alienvault", Subdomain)
		assertResults(t, urls, "alienvault", URL)
	})
}
