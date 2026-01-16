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
	Register(CrtSh)
}

// CrtSh queries the crt.sh certificate transparency database.
var CrtSh = Source{
	Name:   "crtsh",
	Yields: Subdomain,
	Run:    runCrtSh,
}

func runCrtSh(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			yield(Result{}, fmt.Errorf("crtsh: %w", err))
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			yield(Result{}, fmt.Errorf("crtsh: %w", err))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			yield(Result{}, fmt.Errorf("crtsh: unexpected status %d", resp.StatusCode))
			return
		}

		var records []struct {
			NameValue string `json:"name_value"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
			yield(Result{}, fmt.Errorf("crtsh: %w", err))
			return
		}

		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("crtsh: %w", err))
			return
		}

		for _, record := range records {
			// name_value may contain multiple subdomains separated by newlines
			for _, line := range strings.Split(record.NameValue, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				for _, sub := range extractor.Extract(line) {
					if !yield(Result{Type: Subdomain, Value: sub, Source: "crtsh"}, nil) {
						return
					}
				}
			}
		}
	}
}
