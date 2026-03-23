package auth

import (
    "context"
    "net/http"
    "strings"

	"github.com/Justdan111/proxi-api/pkg/response"
)

type contextKey string
const UserIDKey contextKey = "userID"

func (s *Service) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            response.Error(w, http.StatusUnauthorized, "missing authorization header")
            return
        }

        // Expect "Bearer <token>"
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Error(w, http.StatusUnauthorized, "invalid authorization format")
            return
        }

        claims, err := s.ValidateToken(parts[1])
        if err != nil {
            response.Error(w, http.StatusUnauthorized, "invalid or expired token")
            return
        }

        // Attach userID to request context
        ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Helper to get userID from any handler
func GetUserID(r *http.Request) string {
    userID, _ := r.Context().Value(UserIDKey).(string)
    return userID
}