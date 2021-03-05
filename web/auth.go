package web

import (
	"net/http"
	"quote_gallery/database"
)

// adminAuth handles admin authorization
func adminAuth(handler func(w http.ResponseWriter, r *http.Request, u int32)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if ok {
			a := database.IsAdmin(user, password)
			if a != 0 {
				handler(w, r, a)
				return
			}
		}
		// no access granted
		w.Header().Set("WWW-Authenticate", `Basic realm="Log in (admin)"`)
		w.WriteHeader(401)
		w.Write([]byte("You not unauthorized as admin.\n"))
	}
}

// userAuth handles user authorization
func userAuth(handler func(w http.ResponseWriter, r *http.Request, u int32)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if ok {
			u := database.IsUser(user, password)
			if u != 0 {
				handler(w, r, u)
				return
			}
		}
		// no access granted
		w.Header().Set("WWW-Authenticate", `Basic realm="Log in (user)"`)
		w.WriteHeader(401)
		w.Write([]byte("You are not authorized as user.\n"))
	}
}

// anyAuth handles user/admin authorization, passes along isAdmin bool
func anyAuth(handler func(w http.ResponseWriter, r *http.Request, isAdmin bool)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if ok {
			a := database.IsAdmin(user, password)
			if a != 0 {
				handler(w, r, true)
				return
			}
			u := database.IsUser(user, password)
			if u != 0 {
				handler(w, r, false)
				return
			}
		}
		// no access granted
		w.Header().Set("WWW-Authenticate", `Basic realm="Log in (user/admin)"`)
		w.WriteHeader(401)
		w.Write([]byte("You are not authorized as user/admin.\n"))
	}
}
