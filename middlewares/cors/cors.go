// Package cors provides a middleware for CORS related purposes.
package cors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Cors set response headers for CORS related purposes.
func Cors(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PUT, PATCH, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "600")
		c.Writer.Header().Set("Cache-Control", "max-age=600")
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	c.Next()
}
