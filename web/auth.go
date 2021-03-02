package web

import (
	"net/http"
	"quote_gallery/database"
)

// adminAuth handles admin authorization
func adminAuth(handler func(w http.ResponseWriter, r *http.Request, u int32)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		u := database.IsAdmin(user, password)
		if !ok || u == 0 {
			w.Header().Set("WWW-Authenticate", `Basic realm="You need admin priviliges"`)
			w.WriteHeader(401)
			w.Write([]byte("You are unauthorized.\n"))
			return
	  	}
	  	handler(w, r, u)
	}
}

// userAuth handles user authorization
func userAuth(handler func(w http.ResponseWriter, r *http.Request, u int32)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		u := database.IsUser(user, password)
		if !ok || u == 0 {
			w.Header().Set("WWW-Authenticate", `Basic realm="You need user priviliges"`)
			w.WriteHeader(401)
			w.Write([]byte("You are unauthorized.\n"))
			return
	  	}
	  	handler(w, r, u)
	}
}