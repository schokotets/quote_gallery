package web

import (
	"fmt"
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

	rangeBegin := 0

	if pageQuery, ok := r.URL.Query()["page"]; ok {
		page, err := strconv.Atoi(pageQuery[0])
		if err == nil && page >= 0 && len(*quotes) > page*15 {
			rangeBegin = page * 15
		}
	}

	rangeEnd := rangeBegin + 10

	if len(*quotes) < rangeEnd {
		rangeEnd = len(*quotes)
	}

	tmpl := template.Must(template.ParseFiles("pages/quotes.html"))
	tmpl.Execute(w, (*quotes)[rangeBegin:rangeEnd])
}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	teachers, _ := database.GetTeachers()
	tmpl := template.Must(template.ParseFiles("pages/submit.html"))
	tmpl.Execute(w, teachers)
}
