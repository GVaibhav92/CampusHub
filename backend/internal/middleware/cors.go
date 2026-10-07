package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware handling cross-origin requests for local dev and Amplify deployments.
func CORS() gin.HandlerFunc {
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv != "" {
		for _, origin := range strings.Split(allowedOriginsEnv, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check origin allow-list
		allowOrigin := ""
		if origin != "" {
			if len(allowedOrigins) == 0 {
				// Default development allow-list if ALLOWED_ORIGINS env var is not set
				if strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "http://127.0.0.1:") ||
					strings.HasSuffix(origin, ".amplifyapp.com") {
					allowOrigin = origin
				}
			} else {
				for _, allowed := range allowedOrigins {
					if allowed == "*" || allowed == origin {
						allowOrigin = origin
						break
					}
				}
			}
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Origin, X-Requested-With")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Preflight OPTIONS handling
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
