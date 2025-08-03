package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// ChainFunc is a convenience function for http.HandlerFunc
func ChainFunc(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	return Chain(middlewares...)(handler).ServeHTTP
}