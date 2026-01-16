package sources

import (
	"context"
	"fmt"
	"io"
	"iter"
	"net/http"
	"regexp"
)

func init() {
	Register(SiteDossier)
}

// SiteDossier queries the SiteDossier website for subdomains.
var SiteDossier = Source{
	Name:   "sitedossier",
	Yields: Subdomain,
	Run:    runSiteDossier,
}

var siteDossierNextPattern = regexp.MustCompile(`<a href="([A-Za-z0-9/.]+)"><b>`)

func runSiteDossier(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("sitedossier: %w", err))
			return
		}

		// Start with the initial URL
		currentURL := "http://www.sitedossier.com/parentdomain/" + domain

		for currentURL != "" {
			if ctx.Err() != nil {
				return
			}

			body, nextPath, err := fetchSiteDossierPage(ctx, client, currentURL)
			if err != nil {
				yield(Result{}, fmt.Errorf("sitedossier: %w", err))
				return
			}

			// Extract subdomains from page
			for _, sub := range extractor.Extract(body) {
				if !yield(Result{Type: Subdomain, Value: sub, Source: "sitedossier"}, nil) {
					return
				}
			}

			// Follow next link if found
			if nextPath != "" {
				currentURL = "http://www.sitedossier.com" + nextPath
			} else {
				currentURL = ""
			}
		}
	}
}

func fetchSiteDossierPage(ctx context.Context, client *http.Client, url string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	body := string(bodyBytes)

	// Find next page link
	var nextPath string
	if matches := siteDossierNextPattern.FindStringSubmatch(body); len(matches) >= 2 {
		nextPath = matches[1]
	}

	return body, nextPath, nil
}
