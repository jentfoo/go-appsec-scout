package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
)

func init() {
	Register(Anubis)
}

// Anubis queries the Anubis API for subdomains.
var Anubis = Source{
	Name:   "anubis",
	Yields: Subdomain,
	Run:    runAnubis,
}

func runAnubis(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		url := "https://jonlu.ca/anubis/subdomains/" + domain

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			yield(Result{}, fmt.Errorf("anubis: %w", err))
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			yield(Result{}, fmt.Errorf("anubis: %w", err))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			yield(Result{}, fmt.Errorf("anubis: unexpected status %d", resp.StatusCode))
			return
		}

		var subdomains []string
		if err := json.NewDecoder(resp.Body).Decode(&subdomains); err != nil {
			yield(Result{}, fmt.Errorf("anubis: %w", err))
			return
		}

		for _, sub := range subdomains {
			if !yield(Result{Type: Subdomain, Value: sub, Source: "anubis"}, nil) {
				return
			}
		}
	}
}
