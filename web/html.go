package web

import (
	"fmt"
	"html/template"
	"net/http"
	"quote_gallery/database"
	"strconv"
)

//server returns HTML data

const quotesPerPage = 15

func pageRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(404)
		fmt.Fprint(w, "404 Not Found")
		return
	}

	nquotes := database.GetQuotesAmount()
	lastPage := (nquotes-1)/quotesPerPage

	previousPage := -1
	currentPage := 0
	nextPage := 1

	if pageQuery, ok := r.URL.Query()["page"]; ok {
		page, err := strconv.Atoi(pageQuery[0])
		if err == nil && page >= 0 && page*quotesPerPage <= nquotes-1 {
			currentPage = page
			previousPage = page-1
			nextPage = page+1
		}
	}

	if nquotes <= nextPage*quotesPerPage {
		nextPage = -1
	}

	quotes, err := database.GetNQuotesFrom(quotesPerPage, currentPage*quotesPerPage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if quotes == nil {
		quotes, err = database.GetNQuotesFrom(quotesPerPage, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	data := struct {
		Quotes	*[]database.QuoteT
		Prev	int
		Current	int
		Next	int
		Last	int
	}{quotes, previousPage, currentPage, nextPage, lastPage}

	tmpl := template.Must(template.New("quotes.html").Funcs(template.FuncMap{
		"inc": func (i int) int { return i+1 },
	}).ParseFiles("pages/quotes.html"))
	tmpl.Execute(w, data)
}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	teachers, _ := database.GetTeachers()
	tmpl := template.Must(template.ParseFiles("pages/submit.html"))
	tmpl.Execute(w, teachers)
}
