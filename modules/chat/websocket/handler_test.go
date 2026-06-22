package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name         string
		allowed      []string
		isProduction bool
		origin       string
		want         bool
	}{
		{name: "empty origin always allowed (non-browser)", allowed: []string{"https://app.com"}, origin: "", want: true},
		{name: "exact match allowed", allowed: []string{"https://app.com"}, origin: "https://app.com", want: true},
		{name: "suffix attack denied", allowed: []string{"https://app.com"}, origin: "https://app.com.evil.com", want: false},
		{name: "prefix attack denied", allowed: []string{"https://app.com"}, origin: "https://app.com/../evil", want: false},
		{name: "multi allowlist match", allowed: []string{"https://a.com", "https://b.com"}, origin: "https://b.com", want: true},
		{name: "no allowlist denies browser origin in prod", allowed: nil, isProduction: true, origin: "https://x.com", want: false},
		{name: "no allowlist allows browser origin in dev", allowed: nil, isProduction: false, origin: "https://x.com", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{allowedOrigins: tc.allowed, isProduction: tc.isProduction}
			if got := h.isOriginAllowed(tc.origin); got != tc.want {
				t.Errorf("isOriginAllowed(%q) = %v, want %v", tc.origin, got, tc.want)
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newCtx := func(headers map[string]string, query string) *gin.Context {
		req := httptest.NewRequest(http.MethodGet, "/ws"+query, nil)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		return c
	}

	h := &Handler{}

	tests := []struct {
		name    string
		headers map[string]string
		query   string
		want    string
	}{
		{name: "authorization bearer header", headers: map[string]string{"Authorization": "Bearer abc123"}, want: "abc123"},
		{name: "subprotocol bearer", headers: map[string]string{"Sec-WebSocket-Protocol": "bearer, xyz789"}, want: "xyz789"},
		{name: "query fallback", query: "?token=qtok", want: "qtok"},
		{name: "header takes precedence over query", headers: map[string]string{"Authorization": "Bearer htok"}, query: "?token=qtok", want: "htok"},
		{name: "no token", want: ""},
		{name: "malformed authorization ignored falls back to query", headers: map[string]string{"Authorization": "Basic zzz"}, query: "?token=qtok", want: "qtok"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newCtx(tc.headers, tc.query)
			if got := h.extractToken(ctx); got != tc.want {
				t.Errorf("extractToken() = %q, want %q", got, tc.want)
			}
		})
	}
}
