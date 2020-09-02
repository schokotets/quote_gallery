package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"quote_gallery/database"
	//	"strings"
)

type Quote struct {
	Teacher string
	Text    string
}

var quotes []Quote

func handlerMain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "View quotes at /quotes; submit them at /submit")
}

func handlerQuotes(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		quotes = append(quotes, Quote{Teacher: r.FormValue("teacher"), Text: r.FormValue("text")})

	}
	tmpl := template.Must(template.ParseFiles("quotes.html"))
	tmpl.Execute(w, quotes)

}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "submit.html")
}

func main() {
	log.Print("Connecting to database on :5432")
	database.SetupDatabase()
	defer database.CloseDatabase()

	log.Print("Starting website on :8080")
	http.HandleFunc("/", handlerMain)
	http.HandleFunc("/quotes", handlerQuotes)
	http.HandleFunc("/submit", pageSubmit)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
