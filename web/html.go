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

	nquotes := database.GetQuotesAmount()
	rangeBegin := 0
	previousPage := -1
	nextPage := 1

	if pageQuery, ok := r.URL.Query()["page"]; ok {
		page, err := strconv.Atoi(pageQuery[0])
		if err == nil && page >= 0 && page*15 <= nquotes-1 {
			rangeBegin = page * 15
			previousPage = page-1
			nextPage = page+1
		}
	}

	if nquotes <= nextPage*15 {
		nextPage = -1
	}

	quotes, err := database.GetNQuotesFrom(15, rangeBegin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if quotes == nil {
		quotes, err = database.GetNQuotesFrom(15, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	data := struct {
		Quotes	*[]database.QuoteT
		Prev	int
		Next	int
	}{quotes, previousPage, nextPage}

	tmpl := template.Must(template.ParseFiles("pages/quotes.html"))
	tmpl.Execute(w, data)
}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	teachers, _ := database.GetTeachers()
	tmpl := template.Must(template.ParseFiles("pages/submit.html"))
	tmpl.Execute(w, teachers)
}
