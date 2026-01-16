package sources

// TODO: Implement Driftnet source
//
// API Details:
//   Endpoint: https://api.driftnet.io/v1/ (multiple endpoints)
//   Method:   GET
//   Auth:     authorization: Bearer {token} header
//   Yields:   Subdomain
//   Notes:    Query 4 endpoints in parallel (ct/log, scan/protocols, scan/domains, domain/rdns)
