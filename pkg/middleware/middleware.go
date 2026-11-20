// Package middleware berisi middleware Gin yang dipakai lintas modul.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
	"github.com/dedysetiawan/product-catalog-project/pkg/token"
)

const (
	CtxUserID = "user_id"
	CtxEmail  = "user_email"
)

type AdminUserVerifier interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

// JWTAuth memvalidasi header "Authorization: Bearer <token>".
func JWTAuth(tm *token.Manager, users AdminUserVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found || raw == "" {
			response.Error(c, apperror.Unauthorized("missing bearer token"))
			return
		}
		claims, err := tm.Parse(raw)
		if err != nil {
			response.Error(c, apperror.Unauthorized("invalid or expired token"))
			return
		}
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			response.Error(c, apperror.Unauthorized("invalid or expired token"))
			return
		}
		exists, err := users.Exists(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, apperror.Internal(err))
			return
		}
		if !exists {
			response.Error(c, apperror.Unauthorized("invalid or expired token"))
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxEmail, claims.Email)
		c.Next()
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, rec any) {
		slog.Error("panic recovered", "panic", rec)
		response.Error(c, apperror.Internal(nil))
	})
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
		)
	}
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.TrimSpace(o)] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
