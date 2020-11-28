package web

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//SetupRoutes configures which paths are handled by which functions
func SetupRoutes() {
	rt := mux.NewRouter()

	rt.HandleFunc("/", pageRoot)

	handlerFiles := http.FileServer(http.Dir("./public"))
	rt.PathPrefix("/static/").Handler(http.StripPrefix("/static/", handlerFiles))

	rt.HandleFunc("/submit", pageSubmit)
	rt.HandleFunc("/admin/unverifiedquotes", pageAdminUnverifiedQuotes)
	rt.HandleFunc("/admin/unverifiedquotes/{id:[0-9]+}/edit", pageAdminUnverifiedQuotesIDEdit)
	rt.HandleFunc("/api/quotes/submit", handlerAPIQuotesSubmit)
	rt.HandleFunc("/api/unverifiedquotes/{id:[0-9]+}", handlerAPIUnverifiedQuotesID)
	rt.HandleFunc("/api/unverifiedquotes/{id:[0-9]+}/confirm", handlerAPIUnverifiedQuotesIDConfirm)
	rt.HandleFunc("/api/teachers", handlerAPITeachers)
	rt.HandleFunc("/api/teachers/{id:[0-9]+}", handlerAPITeachersID)

	// Direct http handling to gorilla/mux router
	http.Handle("/", rt)
}

//StartWebserver runs the go http.ListenAndServe web server
func StartWebserver() {
	log.Fatal(http.ListenAndServe(":8080", nil))
}
