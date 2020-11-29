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

	// pages
	rt.HandleFunc("/submit", pageSubmit)

	// admin pages
	rt.HandleFunc("/admin", pageAdmin)
	rt.HandleFunc("/admin/unverifiedquotes/{id:[0-9]+}/edit", pageAdminUnverifiedQuotesIDEdit)

	// /api/quotes
	rt.HandleFunc("/api/quotes/submit", postAPIQuotesSubmit).Methods("POST")

	// /api/unverifiedquotes
	rt.HandleFunc("/api/unverifiedquotes/{id:[0-9]+}", putAPIUnverifiedQuotesID).Methods("PUT")
	rt.HandleFunc("/api/unverifiedquotes/{id:[0-9]+}", deleteAPIUnverifiedQuotesID).Methods("DELETE")
	rt.HandleFunc("/api/unverifiedquotes/{id:[0-9]+}/confirm", putAPIUnverifiedQuotesIDConfirm).Methods("PUT")

	// /api/teachers
	rt.HandleFunc("/api/teachers", postAPITeachers).Methods("POST")
	rt.HandleFunc("/api/teachers/{id:[0-9]+}", putAPITeachersID).Methods("PUT")

	// Direct http handling to gorilla/mux router
	http.Handle("/", rt)
}

//StartWebserver runs the go http.ListenAndServe web server
func StartWebserver() {
	log.Fatal(http.ListenAndServe(":8080", nil))
}
