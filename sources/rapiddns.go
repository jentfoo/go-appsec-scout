package sources

import (
	"context"
	"fmt"
	"io"
	"iter"
	"net/http"
	"regexp"
	"strconv"
)

func init() {
	Register(RapidDNS)
}

// RapidDNS queries the RapidDNS website for subdomains.
var RapidDNS = Source{
	Name:   "rapiddns",
	Yields: Subdomain,
	Run:    runRapidDNS,
}

var rapidDNSPagePattern = regexp.MustCompile(`class="page-link"\s+href="/subdomain/[^?]+\?page=(\d+)"`)

func runRapidDNS(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("rapiddns: %w", err))
			return
		}

		// Fetch first page to determine max pages
		body, maxPage, err := fetchRapidDNSPage(ctx, client, domain, 1)
		if err != nil {
			yield(Result{}, fmt.Errorf("rapiddns: %w", err))
			return
		}

		// Extract subdomains from first page
		for _, sub := range extractor.Extract(body) {
			if !yield(Result{Type: Subdomain, Value: sub, Source: "rapiddns"}, nil) {
				return
			}
		}

		// Fetch remaining pages
		for page := 2; page <= maxPage; page++ {
			if ctx.Err() != nil {
				return
			}

			body, _, err := fetchRapidDNSPage(ctx, client, domain, page)
			if err != nil {
				yield(Result{}, fmt.Errorf("rapiddns: page %d: %w", page, err))
				return
			}

			for _, sub := range extractor.Extract(body) {
				if !yield(Result{Type: Subdomain, Value: sub, Source: "rapiddns"}, nil) {
					return
				}
			}
		}
	}
}

func fetchRapidDNSPage(ctx context.Context, client *http.Client, domain string, page int) (string, int, error) {
	url := fmt.Sprintf("https://rapiddns.io/subdomain/%s?page=%d&full=1", domain, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	body := string(bodyBytes)

	// Extract max page from pagination links
	maxPage := 1
	matches := rapidDNSPagePattern.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			if p, err := strconv.Atoi(m[1]); err == nil && p > maxPage {
				maxPage = p
			}
		}
	}

	return body, maxPage, nil
}
