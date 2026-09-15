package utils

import (
	"net/http"
	"net/url"
	"strings"
)

// IsOriginAllowed accepts non-browser clients, same-origin browser requests,
// and explicit origins configured by the operator. Wildcard origins are not
// supported because these websocket endpoints are authenticated control paths.
func IsOriginAllowed(r *http.Request, allowed string) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	if strings.EqualFold(parsed.Host, r.Host) {
		return true
	}
	for _, configured := range strings.Split(allowed, ",") {
		configured = strings.TrimSpace(configured)
		if configured == "" {
			continue
		}
		candidate, err := url.Parse(configured)
		if err == nil && strings.EqualFold(candidate.Scheme, parsed.Scheme) && strings.EqualFold(candidate.Host, parsed.Host) {
			return true
		}
	}
	return false
}
