package middleware

import (
	"context"
	"net/url"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestCORS_SetsHeaders(t *testing.T) {
	cors := CORS()
	c := app.NewContext(0)
	c.Request.SetMethod("GET")
	c.Request.Header.Set("Origin", "http://localhost:3000")

	cors(context.Background(), c)

	origin := string(c.Response.Header.Peek("Access-Control-Allow-Origin"))
	if origin != "http://localhost:3000" {
		t.Errorf("expected origin mirror, got %q", origin)
	}
	methods := string(c.Response.Header.Peek("Access-Control-Allow-Methods"))
	if methods == "" {
		t.Error("expected Allow-Methods header")
	}
	headers := string(c.Response.Header.Peek("Access-Control-Allow-Headers"))
	if headers == "" {
		t.Error("expected Allow-Headers header")
	}
	creds := string(c.Response.Header.Peek("Access-Control-Allow-Credentials"))
	if creds != "true" {
		t.Errorf("expected credentials true, got %q", creds)
	}
	maxAge := string(c.Response.Header.Peek("Access-Control-Max-Age"))
	if maxAge != "86400" {
		t.Errorf("expected max-age 86400, got %q", maxAge)
	}
}

func TestCORS_DefaultOrigin(t *testing.T) {
	cors := CORS()
	c := app.NewContext(0)
	c.Request.SetMethod("GET")

	cors(context.Background(), c)

	origin := string(c.Response.Header.Peek("Access-Control-Allow-Origin"))
	if origin != "*" {
		t.Errorf("expected * when no Origin, got %q", origin)
	}
}

func TestCORS_OptionsReturns204(t *testing.T) {
	cors := CORS()
	c := app.NewContext(0)
	c.Request.SetMethod("OPTIONS")
	c.Request.Header.Set("Origin", "http://localhost:3000")

	cors(context.Background(), c)

	if c.Response.StatusCode() != 204 {
		t.Errorf("expected 204 for OPTIONS, got %d", c.Response.StatusCode())
	}
}

func TestCORS_AllowMethods(t *testing.T) {
	cors := CORS()
	c := app.NewContext(0)
	c.Request.SetMethod("GET")

	cors(context.Background(), c)

	methods := string(c.Response.Header.Peek("Access-Control-Allow-Methods"))
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	for _, m := range expectedMethods {
		if !containsStr(methods, m) {
			t.Errorf("expected %s in Allow-Methods: %q", m, methods)
		}
	}
}

func TestCORS_AllowHeaders(t *testing.T) {
	cors := CORS()
	c := app.NewContext(0)
	c.Request.SetMethod("GET")

	cors(context.Background(), c)

	headers := string(c.Response.Header.Peek("Access-Control-Allow-Headers"))
	expectedHeaders := []string{"Content-Type", "Authorization", "X-Requested-With"}
	for _, h := range expectedHeaders {
		if !containsStr(headers, h) {
			t.Errorf("expected %s in Allow-Headers: %q", h, headers)
		}
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestCORS_OriginSpecialChars(t *testing.T) {
	cors := CORS()
	c := app.NewContext(0)
	c.Request.SetMethod("GET")
	origin := "http://localhost:3000/path?q=1"
	c.Request.Header.Set("Origin", origin)

	cors(context.Background(), c)

	got := string(c.Response.Header.Peek("Access-Control-Allow-Origin"))
	// Should mirror whatever Origin was set
	if got != origin {
		t.Errorf("expected origin %q, got %q", origin, got)
	}
}

func init() {
	_ = url.QueryEscape
}
