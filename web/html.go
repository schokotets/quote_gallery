package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"quote_gallery/database"
	"strconv"
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

func pageAdminUnverifiedQuotes(w http.ResponseWriter, r *http.Request) {
	teachers, err := database.GetUnverifiedQuotes()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get unverified quotes: %v", err)
		return
	}
	tmpl := template.Must(template.ParseFiles("pages/unverifiedquotes.html"))
	tmpl.Execute(w, teachers)
}

func pageAdminUnverifiedQuotesIDEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happen as handlerAPIUnverifiedQuotes is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		return
	}

	quote, err := database.GetUnverifiedQuoteByID(int32(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get unverifiedQuote #%v: %v", id, err)
		return
	}

	teachers, err := database.GetTeachers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get teachers: %v", err)
		return
	}

	editdata := struct {
		Quote database.UnverifiedQuoteT
		Teachers *[]database.TeacherT
	} {
		quote,
		teachers,
	}

	tmpl := template.Must(template.ParseFiles("pages/submit-edit.html"))
	tmpl.Execute(w, editdata)
}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	teachers, err := database.GetTeachers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get teachers")
		return
	}
	tmpl := template.Must(template.ParseFiles("pages/submit.html"))
	tmpl.Execute(w, teachers)
}
