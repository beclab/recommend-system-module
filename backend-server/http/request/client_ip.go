package request

import (
	"net"
	"net/http"
	"strings"
)

func dropIPv6zone(address string) string {
	i := strings.IndexByte(address, '%')
	if i != -1 {
		address = address[:i]
	}
	return address
}

// FindClientIP returns client real IP address.
func FindClientIP(r *http.Request) string {
	headers := []string{"X-Forwarded-For", "X-Real-Ip"}
	for _, header := range headers {
		value := r.Header.Get(header)

		if value != "" {
			addresses := strings.Split(value, ",")
			address := strings.TrimSpace(addresses[0])
			address = dropIPv6zone(address)

			if net.ParseIP(address) != nil {
				return address
			}
		}
	}

	// Fallback to TCP/IP source IP address.
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}
	remoteIP = dropIPv6zone(remoteIP)

	// When listening on a Unix socket, RemoteAddr is empty.
	if remoteIP == "" {
		remoteIP = "127.0.0.1"
	}

	return remoteIP
}
