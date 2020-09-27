package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"quote_gallery/database"
)

func handlerMain(w http.ResponseWriter, r *http.Request) {
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
	tmpl := template.Must(template.ParseFiles("quotes.html"))
	tmpl.Execute(w, quotes)
}

func handlerAPISubmitQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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
	database.StoreQuote(r.FormValue("text"), teacherid)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	teachers, err := database.GetTeachers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles("submit.html"))
	tmpl.Execute(w, teachers)
}

func main() {
	log.Print("Connecting to database on :5432")
	database.SetupDatabase()
	defer database.CloseDatabase()

	log.Print("Starting website on :8080")
	http.HandleFunc("/", handlerMain)
	http.HandleFunc("/api/submitquote", handlerAPISubmitQuote)
	http.HandleFunc("/submit", pageSubmit)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
