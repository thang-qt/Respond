package middleware

import "net/http"

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement per-IP / per-user rate limiting.
		next.ServeHTTP(w, r)
	})
}
