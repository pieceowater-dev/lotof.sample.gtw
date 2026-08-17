package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// NormalizeAPIPathPrefix strips a leading "/api-<service>" segment (e.g.
// "/api-sample/query" -> "/query") so routes registered as plain "/query"
// still match when the gateway sits behind an ALB ingress path-prefix rule
// that does not rewrite the path.
func NormalizeAPIPathPrefix(c *fiber.Ctx) {
	path := c.Path()
	trimmedPath := strings.TrimPrefix(path, "/")
	prefix, remainder, hasRemainder := strings.Cut(trimmedPath, "/")
	if !strings.HasPrefix(prefix, "api-") {
		return
	}

	normalizedPath := "/"
	if hasRemainder {
		normalizedPath += remainder
	}

	c.Path(normalizedPath)
}

// FiberCORSMiddleware handles CORS for Fiber.
func FiberCORSMiddleware(c *fiber.Ctx) {
	origin := c.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	c.Set("Access-Control-Allow-Origin", origin)
	c.Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, SampleAuthorization, Namespace, DeviceId, Fingerprint, X-Requested-With")
	c.Set("Access-Control-Allow-Credentials", "true")
	c.Vary("Origin")
}
