package sources

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register(CommonCrawl)
}

// CommonCrawl queries the Common Crawl index for URLs and subdomains.
var CommonCrawl = Source{
	Name:   "commoncrawl",
	Yields: Subdomain | URL,
	Run:    runCommonCrawl,
}

func runCommonCrawl(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		// Fetch index list
		indexes, err := fetchCommonCrawlIndexes(ctx, client)
		if err != nil {
			yield(Result{}, fmt.Errorf("commoncrawl: %w", err))
			return
		}

		// ByType to last 2 years
		cutoffYear := time.Now().Year() - 2
		var recentIndexes []ccIndex
		for _, idx := range indexes {
			if idx.year >= cutoffYear {
				recentIndexes = append(recentIndexes, idx)
			}
		}

		if len(recentIndexes) == 0 {
			return
		}

		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("commoncrawl: %w", err))
			return
		}

		// Query each index
		for _, idx := range recentIndexes {
			if ctx.Err() != nil {
				return
			}

			endpoint := fmt.Sprintf("%s?url=*.%s&output=text&fl=url", idx.cdxAPI, domain)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if err != nil {
				yield(Result{}, fmt.Errorf("commoncrawl: %w", err))
				return
			}
			req.Header.Set("Host", "index.commoncrawl.org")

			resp, err := client.Do(req)
			if err != nil {
				yield(Result{}, fmt.Errorf("commoncrawl: index %s: %w", idx.id, err))
				return
			}

			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				// Skip this index but continue with others
				continue
			}

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				decoded := decodeCommonCrawlURL(line)
				if decoded == "" {
					continue
				}

				// Yield as URL
				if !yield(Result{Type: URL, Value: decoded, Source: "commoncrawl"}, nil) {
					_ = resp.Body.Close()
					return
				}

				// Extract and yield subdomain
				for _, sub := range extractor.Extract(decoded) {
					if !yield(Result{Type: Subdomain, Value: sub, Source: "commoncrawl"}, nil) {
						_ = resp.Body.Close()
						return
					}
				}
			}

			_ = resp.Body.Close()

			if err := scanner.Err(); err != nil {
				yield(Result{}, fmt.Errorf("commoncrawl: %w", err))
				return
			}
		}
	}
}

type ccIndex struct {
	id     string
	cdxAPI string
	year   int
}

func fetchCommonCrawlIndexes(ctx context.Context, client *http.Client) ([]ccIndex, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://index.commoncrawl.org/collinfo.json", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var raw []struct {
		ID     string `json:"id"`
		CDXAPI string `json:"cdx-api"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var indexes []ccIndex
	for _, r := range raw {
		// Extract year from ID like "CC-MAIN-2024-10"
		year := parseCommonCrawlYear(r.ID)
		if year > 0 {
			indexes = append(indexes, ccIndex{id: r.ID, cdxAPI: r.CDXAPI, year: year})
		}
	}

	return indexes, nil
}

func parseCommonCrawlYear(id string) int {
	// ID format: CC-MAIN-2024-10
	parts := strings.Split(id, "-")
	if len(parts) >= 3 {
		year, err := strconv.Atoi(parts[2])
		if err == nil && year >= 2000 && year <= 2100 {
			return year
		}
	}
	return 0
}

// decodeCommonCrawlURL decodes URL-encoded strings and strips encoding artifacts.
func decodeCommonCrawlURL(raw string) string {
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		decoded = raw
	}

	// Strip common double-encoding artifacts
	decoded = strings.ReplaceAll(decoded, "%25", "%")
	decoded = strings.ReplaceAll(decoded, "%2f", "/")
	decoded = strings.ReplaceAll(decoded, "%2F", "/")

	return decoded
}
