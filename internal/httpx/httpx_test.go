package httpx

import (
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.7:5555"
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")

	if got := ClientIP(r, false); got != "10.0.0.7" {
		t.Errorf("untrusted proxy: got %q, want the socket peer", got)
	}
	// A client-supplied leftmost entry must not be believed.
	if got := ClientIP(r, true); got != "203.0.113.9" {
		t.Errorf("trusted proxy: got %q, want the entry our proxy appended", got)
	}
	r.Header.Set("X-Forwarded-For", "203.0.113.9")
	if got := ClientIP(r, true); got != "203.0.113.9" {
		t.Errorf("single entry: got %q", got)
	}
}
