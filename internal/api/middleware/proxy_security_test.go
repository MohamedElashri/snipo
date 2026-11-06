package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"
)

func TestAppendedProxyAddressesCannotBypassLimits(t *testing.T) {
	old := TrustProxy
	TrustProxy = true
	t.Cleanup(func() { TrustProxy = old })
	limiter := NewRateLimiter(1, time.Minute)
	backend := httptest.NewServer(limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })))
	defer backend.Close()
	destination, _ := url.Parse(backend.URL)
	// Go's reverse proxy appends the observed client IP, as the documented Nginx configuration does.
	proxy := httptest.NewServer(httputil.NewSingleHostReverseProxy(destination))
	defer proxy.Close()
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodPost, proxy.URL+"/login-probe", nil)
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", i+1))
		resp, err := proxy.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
		expected := http.StatusOK
		if i == 1 {
			expected = http.StatusTooManyRequests
		}
		if resp.StatusCode != expected {
			t.Errorf("rotated prefix %d: want %d, got %d", i, expected, resp.StatusCode)
		}
	}
}

func TestForwardedIdentityValidation(t *testing.T) {
	old := TrustProxy
	TrustProxy = true
	t.Cleanup(func() { TrustProxy = old })
	for _, tc := range []struct {
		name      string
		forwarded []string
		real      string
		peer      string
		expected  string
	}{
		{"repeated headers", []string{"198.51.100.8", "198.51.100.9, 203.0.113.10"}, "", "10.0.0.1:8080", "203.0.113.10"},
		{"invalid last hop", []string{"198.51.100.8, not-an-address"}, "192.0.2.9", "10.0.0.1:8080", "10.0.0.1"},
		{"empty last hop", []string{"198.51.100.8,"}, "", "10.0.0.1:8080", "10.0.0.1"},
		{"IPv6", []string{"2001:0db8:0000::1"}, "", "10.0.0.1:8080", "2001:db8::1"},
		{"mapped IPv4", []string{"::ffff:192.0.2.10"}, "", "10.0.0.1:8080", "192.0.2.10"},
		{"real IP only", nil, "203.0.113.11", "10.0.0.1:8080", "203.0.113.11"},
		{"invalid real IP", nil, "untrusted-label", "10.0.0.1:8080", "10.0.0.1"},
		{"direct IPv6", nil, "", "[2001:0db8::1]:8080", "2001:db8::1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.peer
			for _, value := range tc.forwarded {
				req.Header.Add("X-Forwarded-For", value)
			}
			if tc.real != "" {
				req.Header.Set("X-Real-IP", tc.real)
			}
			if got := ClientIP(req); got != tc.expected {
				t.Errorf("identity=%q, want %q", got, tc.expected)
			}
		})
	}
	TrustProxy = false
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:9000"
	req.Header.Set("X-Forwarded-For", "198.51.100.1")
	req.Header.Set("X-Real-IP", "198.51.100.2")
	if got := ClientIP(req); got != "192.0.2.1" {
		t.Errorf("untrusted headers affected direct identity: %q", got)
	}
}
