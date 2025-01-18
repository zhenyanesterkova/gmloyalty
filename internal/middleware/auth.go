package middleware

import (
	"context"
	"net/http"
)

type contextKey uint

const (
	UserIDContextKey contextKey = iota
)

var (
	noAuthUrls = map[string]struct{}{
		"/api/user/register": struct{}{},
		"/api/user/login":    struct{}{},
	}
)

func (lm MiddlewareStruct) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := noAuthUrls[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		tokenJWT := r.Header.Get("Authorization")

		userID, err := lm.jwtSess.Check(tokenJWT)
		if err != nil {
			http.Error(w, "No auth", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
