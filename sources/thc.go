package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
)

func init() {
	Register(THC)
}

// THC queries the THC API for subdomains.
var THC = Source{
	Name:   "thc",
	Yields: Subdomain,
	Run:    runTHC,
}

type thcRequest struct {
	Domain    string `json:"domain"`
	PageState string `json:"page_state"`
	Limit     int    `json:"limit"`
}

type thcResponse struct {
	Domains []struct {
		Domain string `json:"domain"`
	} `json:"domains"`
	NextPageState string `json:"next_page_state"`
}

func runTHC(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		pageState := ""

		for {
			if ctx.Err() != nil {
				return
			}

			reqBody := thcRequest{
				Domain:    domain,
				PageState: pageState,
				Limit:     1000,
			}

			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				yield(Result{}, fmt.Errorf("thc: %w", err))
				return
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://ip.thc.org/api/v1/lookup/subdomains", bytes.NewReader(bodyBytes))
			if err != nil {
				yield(Result{}, fmt.Errorf("thc: %w", err))
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				yield(Result{}, fmt.Errorf("thc: %w", err))
				return
			}

			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				yield(Result{}, fmt.Errorf("thc: unexpected status %d", resp.StatusCode))
				return
			}

			var response thcResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				_ = resp.Body.Close()
				yield(Result{}, fmt.Errorf("thc: %w", err))
				return
			}
			_ = resp.Body.Close()

			// Yield domains from this page
			for _, d := range response.Domains {
				if d.Domain == "" {
					continue
				}
				if !yield(Result{Type: Subdomain, Value: d.Domain, Source: "thc"}, nil) {
					return
				}
			}

			// Check for more pages
			if response.NextPageState == "" {
				return
			}
			pageState = response.NextPageState
		}
	}
}
