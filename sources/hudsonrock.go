package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"strings"
)

func init() {
	Register(HudsonRock)
}

// HudsonRock queries the HudsonRock API for breach data URLs.
var HudsonRock = Source{
	Name:   "hudsonrock",
	Yields: Subdomain | URL,
	Run:    runHudsonRock,
}

func runHudsonRock(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		endpoint := "https://cavalier.hudsonrock.com/api/json/v2/osint-tools/urls-by-domain?domain=" + domain

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			yield(Result{}, fmt.Errorf("hudsonrock: %w", err))
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			yield(Result{}, fmt.Errorf("hudsonrock: %w", err))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			yield(Result{}, fmt.Errorf("hudsonrock: unexpected status %d", resp.StatusCode))
			return
		}

		var response struct {
			Data struct {
				EmployeesURLs []struct {
					URL string `json:"url"`
				} `json:"employees_urls"`
				ClientsURLs []struct {
					URL string `json:"url"`
				} `json:"clients_urls"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			yield(Result{}, fmt.Errorf("hudsonrock: %w", err))
			return
		}

		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("hudsonrock: %w", err))
			return
		}

		// Process both employees and clients URLs
		allURLs := make([]string, 0, len(response.Data.EmployeesURLs)+len(response.Data.ClientsURLs))
		for _, u := range response.Data.EmployeesURLs {
			allURLs = append(allURLs, u.URL)
		}
		for _, u := range response.Data.ClientsURLs {
			allURLs = append(allURLs, u.URL)
		}

		for _, u := range allURLs {
			u = strings.TrimSpace(u)
			if u == "" {
				continue
			}

			// Yield as URL
			if !yield(Result{Type: URL, Value: u, Source: "hudsonrock"}, nil) {
				return
			}

			// Extract and yield subdomain
			for _, sub := range extractor.Extract(u) {
				if !yield(Result{Type: Subdomain, Value: sub, Source: "hudsonrock"}, nil) {
					return
				}
			}
		}
	}
}
