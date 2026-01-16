package sources

// TODO: Implement VirusTotal source
//
// API Details:
//   Subdomain Endpoint: https://www.virustotal.com/api/v3/domains/{domain}/subdomains?limit=40&cursor={cursor}
//   URL Endpoint:       https://www.virustotal.com/vtapi/v2/domain/report?apikey={key}&domain={domain}
//   Method:             GET
//   Auth:               x-apikey header (v3) or query param (v2)
//   Yields:             Subdomain, URL
//   Notes:              Cursor-based pagination for v3
