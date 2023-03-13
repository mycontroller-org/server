package http_handler

import "strings"

// ExtractHost splits a host / port pair (or just a host) and returns the host.
// This is large borrowed from `net/url.splitHostPort`.
func ExtractHost(hostport string) string {
	host := hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 {
		host = host[:colon]
	}

	// If `hostport` is an IPv6 address of the form `[::1]:12801`.
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return host
}
