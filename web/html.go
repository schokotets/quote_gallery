package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"quote_gallery/database"
	"sort"
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
	tmpl := template.Must(template.New("quotes.html").Funcs(template.FuncMap{
		"GetTeacherByID": database.GetTeacherByID,
	}).ParseFiles("pages/quotes.html"))
	tmpl.Execute(w, quotes)
}

func pageAdmin(w http.ResponseWriter, r *http.Request) {
	quotes, err := database.GetUnverifiedQuotes()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get unverified quotes: %v", err)
		return
	}

	teachers, err := database.GetTeachers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get teachers: %v", err)
		return
	}

	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Unixtime < quotes[j].Unixtime })
	sort.Slice(teachers, func(i, j int) bool { return teachers[i].TeacherID < teachers[j].TeacherID })

	pagedata := struct {
		Quotes []database.UnverifiedQuoteT
		Teachers []database.TeacherT
	} {
		quotes,
		teachers,
	}

	tmpl := template.Must(template.New("admin.html").Funcs(template.FuncMap{
		"GetTeacherByID": database.GetTeacherByID,
	}).ParseFiles("pages/admin.html"))

	err = tmpl.Execute(w, pagedata)
	if err != nil {
		panic(err)
	}
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
		Teachers []database.TeacherT
	} {
		quote,
		teachers,
	}

	tmpl := template.Must(template.ParseFiles("pages/edit-unverifiedquote.html"))
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
