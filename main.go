package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"quote_gallery/database"
	"strconv"
	"time"
)

func handlerMain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "View quotes at /quotes; submit them at /submit")
}

func handlerQuotes(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		teacherid, err := strconv.Atoi(r.FormValue("teacherid"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Form field teacherid is not an integer")
			return
		}
		database.CreateQuote(database.QuoteT{
			Text:      r.FormValue("text"),
			Context:   r.FormValue("context"),
			TeacherID: uint32(teacherid),
			Unixtime:  uint64(time.Now().Unix())})
	}
	quotes := database.GetQuotes()
	tmpl := template.Must(template.ParseFiles("quotes.html"))
	tmpl.Execute(w, quotes)
}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	teachers := database.GetTeachers()
	tmpl := template.Must(template.ParseFiles("submit.html"))
	tmpl.Execute(w, teachers)
}

func main() {
	log.Print("Connecting to database on :5432")
	database.Setup()
	defer database.CloseAndClearCache()

	log.Print("Starting website on :8080")
	http.HandleFunc("/", handlerMain)
	http.HandleFunc("/quotes", handlerQuotes)
	http.HandleFunc("/submit", pageSubmit)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
