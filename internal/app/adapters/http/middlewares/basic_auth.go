package middlewares

import stdhttp "net/http"

func BasicAuth(username string, password string) func(next stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			reqUser, reqPass, ok := r.BasicAuth()
			if !ok || reqUser != username || reqPass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="xray-exporter"`)
				stdhttp.Error(w, "Unauthorized", stdhttp.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
