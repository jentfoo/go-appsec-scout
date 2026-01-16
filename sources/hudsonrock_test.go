package sources

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHudsonRock(t *testing.T) {
	t.Parallel()

	t.Run("registered", func(t *testing.T) {
		src := ByName("hudsonrock")
		require.NotNil(t, src)
		assert.Equal(t, Subdomain|URL, src.Yields)
	})

	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test")
		}

		ctx := t.Context()
		client := &http.Client{Timeout: 30 * time.Second}
		subdomains, urls, errors := collectResults(HudsonRock.Run(ctx, client, "github.com", ""))

		if len(errors) > 0 {
			t.Logf("errors: %v", errors)
		}

		// HudsonRock may not have breach data for all domains
		t.Logf("found %d subdomains, %d urls", len(subdomains), len(urls))
		assert.NotEmpty(t, urls)
		assertResults(t, subdomains, "hudsonrock", Subdomain)
		assertResults(t, urls, "hudsonrock", URL)
	})
}
