package middleware

import "github.com/gofiber/fiber/v2"

// FiberCORSMiddleware handles CORS for Fiber.
func FiberCORSMiddleware(c *fiber.Ctx) {
	origin := c.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	c.Set("Access-Control-Allow-Origin", origin)
	c.Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Namespace, DeviceId, Fingerprint, X-Requested-With")
	c.Set("Access-Control-Allow-Credentials", "true")
	c.Vary("Origin")
}
