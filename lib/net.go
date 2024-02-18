package lib

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

func ResolveClientIP(r *http.Request) (string, error) {
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		addrs := strings.Split(prior, ",")
		if len(addrs) > 0 {
			return addrs[0], nil
		}
	}
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("split host and port: %w", err)
	}
	return clientIP, nil
}
