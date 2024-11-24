package middleware

import "net/http"

func BadRequestIfMethodNotEqual(method string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				http.Error(w, `Bad Request`, http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func BadRequestIfMethodNotEqualPOST(next http.Handler) http.Handler {
	return BadRequestIfMethodNotEqual(http.MethodPost)(next)
}
