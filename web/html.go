package web

import (
	"fmt"
	"html/template"
	"net/http"
	"quote_gallery/database"
)

//server returns HTML data

func pageRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(404)
		fmt.Fprint(w, "404 Not Found")
		return
	}
	quotes, err := database.GetQuotes()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	quotes, _ = database.GetQuotes()
	tmpl := template.Must(template.ParseFiles("pages/quotes.html"))
	tmpl.Execute(w, quotes)
}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	teachers, _ := database.GetTeachers()
	tmpl := template.Must(template.ParseFiles("pages/submit.html"))
	tmpl.Execute(w, teachers)
}
