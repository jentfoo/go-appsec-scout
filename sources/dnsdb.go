package sources

// TODO: Implement DNSDB source
//
// API Details:
//   Endpoint: https://api.dnsdb.info/dnsdb/v2/lookup/rrset/name/*.{domain}
//   Method:   GET
//   Auth:     X-API-KEY header
//   Yields:   Subdomain
//   Notes:    NDJSON response, offset-based pagination
