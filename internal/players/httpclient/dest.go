package httpclient

import (
	"context"
	"net"
	"net/url"
	"strings"

	"github.com/gspaim/Runtgine/internal/core/result"
)

const metadataGoogle = "metadata.google.internal"

var (
	linkLocalV4 = mustCIDR("169.254.0.0/16")
	awsMetaV6   = mustCIDR("fd00:ec2::/128")
)

func mustCIDR(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return n
}

func validateRawURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return result.Validation(result.CodeInvalidInput, "url is required", nil)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid url: "+err.Error(), nil)
	}
	return validateURL(u)
}

func validateURL(u *url.URL) error {
	if u == nil {
		return result.Validation(result.CodeInvalidInput, "url is required", nil)
	}
	if strings.ToLower(u.Scheme) != "https" {
		return result.Validation(result.CodeInvalidInput, "url must use https", map[string]any{"scheme": u.Scheme})
	}
	if u.User != nil {
		return result.Validation(result.CodeInvalidInput, "url must not include userinfo", nil)
	}
	host := u.Hostname()
	if host == "" {
		return result.Validation(result.CodeInvalidInput, "url host is required", nil)
	}
	if forbiddenHost(host) {
		return result.Validation(result.CodeInvalidInput, "destination host is not allowed", map[string]any{"host": host})
	}
	if ip := net.ParseIP(host); ip != nil && forbiddenIP(ip) {
		return result.Validation(result.CodeInvalidInput, "destination is not allowed", map[string]any{"host": host})
	}
	return nil
}

func forbiddenHost(host string) bool {
	h := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	return h == metadataGoogle
}

func forbiddenIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLinkLocalUnicast() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		return linkLocalV4.Contains(ip4)
	}
	return awsMetaV6.Contains(ip)
}

type lookupFunc func(ctx context.Context, host string) ([]net.IP, error)

func defaultLookup(ctx context.Context, host string) ([]net.IP, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	out := make([]net.IP, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, a.IP)
	}
	return out, nil
}

func resolveAllowed(ctx context.Context, host string, lookup lookupFunc) ([]net.IP, error) {
	if lookup == nil {
		lookup = defaultLookup
	}
	if ip := net.ParseIP(host); ip != nil {
		if forbiddenIP(ip) {
			return nil, result.Runtime(result.CodePlayerError, "destination is not allowed", false, map[string]any{"host": host})
		}
		return []net.IP{ip}, nil
	}
	if forbiddenHost(host) {
		return nil, result.Runtime(result.CodePlayerError, "destination host is not allowed", false, map[string]any{"host": host})
	}
	ips, err := lookup(ctx, host)
	if err != nil {
		return nil, result.Runtime(result.CodePlayerError, "dns lookup: "+err.Error(), false, map[string]any{"host": host})
	}
	allowed := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if forbiddenIP(ip) {
			continue
		}
		allowed = append(allowed, ip)
	}
	if len(allowed) == 0 {
		return nil, result.Runtime(result.CodePlayerError, "destination is not allowed", false, map[string]any{"host": host})
	}
	return allowed, nil
}
