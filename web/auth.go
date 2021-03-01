package web

import (
	"net/http"
	"quote_gallery/database"
)

// adminAuth handles admin authorization
func adminAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || database.IsAdmin(user, password) == 0 {
			w.Header().Set("WWW-Authenticate", `Basic realm="You need admin priviliges"`)
			w.WriteHeader(401)
			w.Write([]byte("You are unauthorized.\n"))
			return
	  	}
	  	handler(w, r)
	}
}

// userAuth handles user authorization
func userAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || database.IsUser(user, password) == 0 {
			w.Header().Set("WWW-Authenticate", `Basic realm="You need user priviliges"`)
			w.WriteHeader(401)
			w.Write([]byte("You are unauthorized.\n"))
			return
	  	}
	  	handler(w, r)
	}
}