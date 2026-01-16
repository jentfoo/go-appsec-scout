package sources

// TODO: Implement Facebook CT source
//
// API Details:
//   Endpoint: https://graph.facebook.com/certificates?fields=domains&access_token={token}&query={domain}
//   Method:   GET
//   Auth:     OAuth access token (from app_id + secret)
//   Yields:   Subdomain
//   Notes:    Rate limit ~20,000 req/hour per appID
