package handler

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func trustedClientIP(r *http.Request) string {
	remoteIP := remoteRequestIP(r.RemoteAddr)
	if remoteIP != nil && remoteTrusted(remoteIP) {
		if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
			parts := strings.Split(forwarded, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				ip := net.ParseIP(strings.TrimSpace(parts[i]))
				if ip == nil {
					continue
				}
				if !remoteTrusted(ip) {
					return ip.String()
				}
			}
		}
	}
	if remoteIP != nil {
		return remoteIP.String()
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func publicAPIBaseURL(r *http.Request) string {
	if configured := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_API_BASE_URL")), "/"); configured != "" {
		return configured
	}

	scheme := "http"
	host := r.Host
	if r.TLS != nil {
		scheme = "https"
	}
	remoteIP := remoteRequestIP(r.RemoteAddr)
	if remoteIP != nil && remoteTrusted(remoteIP) {
		if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto == "https" || proto == "http" {
			scheme = proto
		}
		if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
			host = strings.Split(forwardedHost, ",")[0]
		}
	}
	if strings.TrimSpace(host) == "" {
		return ""
	}
	return scheme + "://" + host
}

func joinBaseURL(base, value string) string {
	if value == "" || strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	if base == "" {
		return value
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return value
	}
	if strings.HasPrefix(value, "/") {
		return strings.TrimRight(base, "/") + value
	}
	return strings.TrimRight(base, "/") + "/" + value
}

func remoteRequestIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		host = strings.TrimSpace(remoteAddr)
	}
	return net.ParseIP(host)
}

func remoteTrusted(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, raw := range strings.Split(os.Getenv("TRUSTED_PROXY_CIDRS"), ",") {
		_, network, err := net.ParseCIDR(strings.TrimSpace(raw))
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}
