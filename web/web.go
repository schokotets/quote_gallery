package web

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//SetupRoutes configures which paths are handled by which functions
func SetupRoutes() {
	rt := mux.NewRouter()

	rt.HandleFunc("/", anyAuth(pageRoot))

	// TODO cache svgs longer
	handlerFiles := http.FileServer(http.Dir("./public"))
	rt.PathPrefix("/static/").Handler(http.StripPrefix("/static/", handlerFiles))

	// pages
	rt.HandleFunc("/submit", userAuth(pageSubmit) )
	rt.HandleFunc("/suggestions", userAuth(pageSimilarQuotes) )

	// admin pages
	rt.HandleFunc("/admin", adminAuth(pageAdmin) )
	rt.HandleFunc("/admin/unverifiedquotes/{id:[0-9]+}/edit", adminAuth(pageAdminUnverifiedQuotesIDEdit) )
	rt.HandleFunc("/admin/teachers/{id:[0-9]+}/edit", adminAuth(pageAdminTeachersIDEdit) )
	rt.HandleFunc("/admin/teachers/add", adminAuth(pageAdminTeachersAdd) )

	// /api/quotes
	rt.HandleFunc("/api/quotes/submit", userAuth(postAPIQuotesSubmit) ).Methods("POST")
	rt.HandleFunc("/api/quotes/{id:[0-9]+}/vote/{val:[1-5]}", userAuth(putAPIQuotesIDVoteRating) ).Methods("PUT")

	// /api/unverifiedquotes
	rt.HandleFunc("/api/unverifiedquotes/{id:[0-9]+}", adminAuth(putAPIUnverifiedQuotesID) ).Methods("PUT")
	rt.HandleFunc("/api/unverifiedquotes/{id:[0-9]+}", adminAuth(deleteAPIUnverifiedQuotesID) ).Methods("DELETE")
	rt.HandleFunc("/api/unverifiedquotes/{id:[0-9]+}/confirm", adminAuth(putAPIUnverifiedQuotesIDConfirm) ).Methods("PUT")
	rt.HandleFunc("/api/unverifiedquotes/{quoteid:[0-9]+}/assignteacher/{teacherid:[0-9]+}", adminAuth(putAPIUnverifiedQuotesIDAssignTeacherID)).Methods("PUT")

	// /api/teachers
	rt.HandleFunc("/api/teachers", adminAuth(postAPITeachers) ).Methods("POST")
	rt.HandleFunc("/api/teachers/{id:[0-9]+}", adminAuth(putAPITeachersID) ).Methods("PUT")

	// Direct http handling to gorilla/mux router
	http.Handle("/", rt)
}

//StartWebserver runs the go http.ListenAndServe web server
func StartWebserver() {
	log.Fatal(http.ListenAndServe(":8080", nil))
}
