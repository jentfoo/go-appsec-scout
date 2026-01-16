package sources

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"net/http"
	"strings"
)

func init() {
	Register(HackerTarget)
}

// HackerTarget queries the HackerTarget API for subdomains.
// Works without API key but has rate limits; key improves limits.
var HackerTarget = Source{
	Name:   "hackertarget",
	Yields: Subdomain,
	Run:    runHackerTarget,
}

func runHackerTarget(ctx context.Context, client *http.Client, domain string, apiKey string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		url := "https://api.hackertarget.com/hostsearch/?q=" + domain
		if apiKey != "" {
			url += "&apikey=" + apiKey
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			yield(Result{}, fmt.Errorf("hackertarget: %w", err))
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			yield(Result{}, fmt.Errorf("hackertarget: %w", err))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			yield(Result{}, fmt.Errorf("hackertarget: unexpected status %d", resp.StatusCode))
			return
		}

		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("hackertarget: %w", err))
			return
		}

		// Response is CSV-like: subdomain,ip per line
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// Extract subdomain from line (format: subdomain,ip)
			for _, sub := range extractor.Extract(line) {
				if !yield(Result{Type: Subdomain, Value: sub, Source: "hackertarget"}, nil) {
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			yield(Result{}, fmt.Errorf("hackertarget: %w", err))
		}
	}
}
