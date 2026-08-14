package middleware

import (
	"context"
	"net/url"

	"github.com/cloudwego/hertz/pkg/app"
)

// isLocalOrigin reports whether the origin is a loopback / private-network
// host (the legitimate browser clients for this service: the embedded dev
// server and the Wails desktop webview).
func isLocalOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	switch host {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0", "":
		return true
	}
	// Wails desktop webviews use wails:// or the app's own scheme.
	if u.Scheme == "wails" {
		return true
	}
	return false
}

func CORS() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		if origin != "" && isLocalOrigin(origin) {
			// Reflect only known-local origins so credentialed cross-origin
			// access from arbitrary websites is not possible.
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		c.Header("Access-Control-Max-Age", "86400")

		if string(c.Method()) == "OPTIONS" {
			c.Status(204)
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}
