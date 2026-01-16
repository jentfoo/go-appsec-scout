package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
)

func init() {
	Register(AlienVault)
}

// AlienVault queries the AlienVault OTX URL list endpoint.
var AlienVault = Source{
	Name:   "alienvault",
	Yields: Subdomain | URL,
	Run:    runAlienVault,
}

func runAlienVault(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("alienvault: %w", err))
			return
		}

		page := 1
		for {
			if ctx.Err() != nil {
				return
			}

			url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/url_list?page=%d", domain, page)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				yield(Result{}, fmt.Errorf("alienvault: %w", err))
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				yield(Result{}, fmt.Errorf("alienvault: %w", err))
				return
			}

			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				yield(Result{}, fmt.Errorf("alienvault: unexpected status %d", resp.StatusCode))
				return
			}

			var response struct {
				URLList []struct {
					URL string `json:"url"`
				} `json:"url_list"`
				HasNext bool `json:"has_next"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				_ = resp.Body.Close()
				yield(Result{}, fmt.Errorf("alienvault: %w", err))
				return
			}
			_ = resp.Body.Close()

			// Yield URLs and extract subdomains
			for _, u := range response.URLList {
				if u.URL == "" {
					continue
				}

				// Yield as URL
				if !yield(Result{Type: URL, Value: u.URL, Source: "alienvault"}, nil) {
					return
				}

				// Extract and yield subdomain
				for _, sub := range extractor.Extract(u.URL) {
					if !yield(Result{Type: Subdomain, Value: sub, Source: "alienvault"}, nil) {
						return
					}
				}
			}

			// Check for more pages
			if !response.HasNext {
				return
			}
			page++
		}
	}
}
