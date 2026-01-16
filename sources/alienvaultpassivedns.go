package sources

// TODO: Implement AlienVault OTX PassiveDNS source (auth required)
//
// API Details:
//   Endpoint: https://otx.alienvault.com/api/v1/indicators/domain/{domain}/passive_dns
//   Method:   GET
//   Auth:     Bearer token
//   Yields:   Subdomain
//   Notes:    Different from alienvault.go which uses the URL list endpoint (no auth required)
