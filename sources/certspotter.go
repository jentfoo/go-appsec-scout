package sources

// TODO: Implement CertSpotter source
//
// API Details:
//   Endpoint: https://api.certspotter.com/v1/issuances?domain={domain}&include_subdomains=true&expand=dns_names
//   Method:   GET
//   Auth:     Bearer token
//   Yields:   Subdomain
//   Notes:    Cursor-based pagination using `after` parameter
