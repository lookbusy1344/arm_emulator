package api

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// isLoopbackHost reports whether host (without port) names the local machine.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.Equal(net.IPv4(127, 0, 0, 1)) || ip.Equal(net.IPv6loopback))
}

// isAllowedOrigin reports whether a browser Origin header names a page served from
// this machine or a local file. An empty origin means a non-browser client.
func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.User != nil {
		return false
	}
	if u.Scheme == "file" {
		return u.Host == ""
	}
	if u.Path != "" {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return isLoopbackHost(u.Hostname())
}

// isAllowedRequestOrigin is the websocket.Upgrader CheckOrigin callback.
func isAllowedRequestOrigin(r *http.Request) bool {
	return isAllowedOrigin(r.Header.Get("Origin"))
}

// LoopbackHostOnly rejects requests whose Host header does not name the local machine.
// This blocks DNS rebinding, where a remote page resolves its own hostname to 127.0.0.1
// and then talks to the API as a same-origin client.
func LoopbackHostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if !isLoopbackHost(host) {
			writeError(w, http.StatusForbidden, "Host not allowed")
			return
		}
		next.ServeHTTP(w, r)
	})
}
