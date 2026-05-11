// Package middleware provides HTTP middleware implementations.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"spark-s3-microservice/pkg/firebase"
)

type contextKey string

const UserIDKey contextKey = "uid"

// FirebaseAuth creates a middleware that validates Firebase Bearer tokens.
func FirebaseAuth(verifier firebase.TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
			return
		}

		token := parts[1]

		// Verify token with Firebase
		uid, err := verifier.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			return
		}

		// Set uid in context for handlers
		c.Set(string(UserIDKey), uid)
		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context.
func GetUserID(c *gin.Context) (string, bool) {
	uid, exists := c.Get(string(UserIDKey))
	if !exists {
		return "", false
	}
	return uid.(string), true
}