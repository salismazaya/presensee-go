package middleware

import (
	"context"
	"net/http"
	"strings"

	"presensee/internal/model"

	"gorm.io/gorm"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

func AuthBearer(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""

			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else if qToken := r.URL.Query().Get("token"); qToken != "" {
				token = qToken
			} else if cookie, err := r.Cookie("token"); err == nil {
				token = cookie.Value
			}

			if token == "" {
				http.Error(w, `{"detail":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			var user model.User
			if err := db.Where("token = ? AND is_active = ?", token, true).First(&user).Error; err != nil {
				http.Error(w, `{"detail":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) *model.User {
	if u, ok := ctx.Value(UserContextKey).(*model.User); ok {
		return u
	}
	return nil
}

func RequireRole(roles ...model.UserType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, `{"detail":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			if user.IsSuperuser {
				next.ServeHTTP(w, r)
				return
			}

			if user.Type == nil {
				http.Error(w, `{"detail":"Forbidden"}`, http.StatusForbidden)
				return
			}

			for _, role := range roles {
				if *user.Type == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"detail":"Forbidden"}`, http.StatusForbidden)
		})
	}
}
