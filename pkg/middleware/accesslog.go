package middleware

import (
	"net/http"
)

// WrapAccesslog is a middleware that writes the access log.
func WrapAccesslog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
		h.ServeHTTP(w, r)
	})
}
