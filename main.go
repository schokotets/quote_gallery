package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

var quotes = []string{"Dieser Junge hat keine Ahnung von go,", "aber das ist okay."}

func handlerMain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "View quotes at /quotes; submit them at /submit")
}

func handlerQuotes(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
		    fmt.Fprintf(w, "ParseForm() err: %v", err)
		    return
		}

		teacher := r.FormValue("teacher")
		text := r.FormValue("text")
		fmt.Fprintf(w, "teacher = %s\n", teacher)
		fmt.Fprintf(w, "text = %s\n", text)
	}

	fmt.Fprintf(w, "\nHi there, I love quotes!\n\n")
	fmt.Fprintf(w, strings.Join(quotes, "\n"))

}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "submit.html")
}



func main() {
	log.Print("Starting website on :8080")
	http.HandleFunc("/", handlerMain)
	http.HandleFunc("/quotes", handlerQuotes)
	http.HandleFunc("/submit", pageSubmit)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
