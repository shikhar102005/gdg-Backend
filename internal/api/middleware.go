package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/auth"
)

type contextKey string

const (
	userClaimsKey contextKey = "userClaims"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":     message,
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// Don't log spam for health check polling
		if r.URL.Path != "/healthz" && r.URL.Path != "/readyz" {
			log.Printf("[%s] %s %s - %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
		}
	})
}

func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if allowed {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Sensor-Key")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(authSvc *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := auth.ExtractTokenFromRequest(r)
			if token == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization token")
				return
			}

			claims, err := authSvc.ValidateToken(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...auth.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claimsVal := r.Context().Value(userClaimsKey)
			if claimsVal == nil {
				writeError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}

			claims, ok := claimsVal.(*auth.OrganizerClaims)
			if !ok {
				writeError(w, http.StatusUnauthorized, "invalid user claims")
				return
			}

			// Admin role always has full access
			if claims.Role == auth.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			allowed := false
			for _, rRole := range roles {
				if claims.Role == rRole {
					allowed = true
					break
				}
			}

			if !allowed {
				writeError(w, http.StatusForbidden, "insufficient permissions for this action")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserClaims(r *http.Request) *auth.OrganizerClaims {
	if val := r.Context().Value(userClaimsKey); val != nil {
		if claims, ok := val.(*auth.OrganizerClaims); ok {
			return claims
		}
	}
	return nil
}
