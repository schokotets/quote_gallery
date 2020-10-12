package web

import (
	"log"
	"net/http"
)

//SetupRoutes configures which paths are handled by which functions
func SetupRoutes() {
	http.HandleFunc("/", pageRoot)

	handlerFiles := http.FileServer(http.Dir("./public"))
	http.Handle("/static/", http.StripPrefix("/static/", handlerFiles))

	http.HandleFunc("/submit", pageSubmit)

	http.HandleFunc("/api/quotes/submit", handlerAPIQuotesSubmit)
}

//StartWebserver runs the go http.ListenAndServe web server
func StartWebserver() {
	log.Fatal(http.ListenAndServe(":8080", nil))
}
