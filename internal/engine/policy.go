package engine

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

type Policy struct {
	AllowPrivate bool
	AllowWrite   bool
	AllowedHosts map[string]bool
	AllowedPorts map[string]bool
}

func (p Policy) ValidateMethod(method, path string) error {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return nil
	case "POST", "PUT", "PATCH", "DELETE":
		if !p.AllowWrite {
			return fmt.Errorf("blocked by safety policy: %s requires --allow-write", method)
		}
		for _, dangerous := range []string{"/delete", "/destroy", "/reset", "/drop"} {
			if strings.Contains(strings.ToLower(path), dangerous) {
				return fmt.Errorf("destructive-looking path requires a separate, manually reviewed request")
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported HTTP method %q", method)
	}
}

func (p Policy) ValidateURL(ctx context.Context, rawURL string) ([]string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("TARGET_BLOCKED: invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("TARGET_BLOCKED: unsupported scheme")
	}
	if u.User != nil || u.Hostname() == "" {
		return nil, fmt.Errorf("TARGET_BLOCKED: userinfo and empty hosts are not allowed")
	}
	host := strings.ToLower(u.Hostname())
	if len(p.AllowedHosts) > 0 && !p.AllowedHosts[host] {
		return nil, fmt.Errorf("TARGET_BLOCKED: host is not allowlisted")
	}
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	if len(p.AllowedPorts) > 0 && !p.AllowedPorts[port] {
		return nil, fmt.Errorf("TARGET_BLOCKED: port is not allowlisted")
	}
	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("DNS_FAILURE: %w", err)
	}
	resolved := make([]string, 0, len(addresses))
	for _, address := range addresses {
		address = address.Unmap()
		if !p.AllowPrivate && blockedAddress(address) {
			return nil, fmt.Errorf("TARGET_BLOCKED: %s resolves to a non-public address", host)
		}
		resolved = append(resolved, address.String())
	}
	return resolved, nil
}

func blockedAddress(address netip.Addr) bool {
	if !address.IsValid() {
		return true
	}
	if address.IsLoopback() || address.IsPrivate() || address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsMulticast() || address.IsUnspecified() {
		return true
	}
	metadata := netip.MustParseAddr("169.254.169.254")
	return address == metadata
}
