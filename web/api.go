package web

import (
	"fmt"
	"net/http"
	"quote_gallery/database"
	"strconv"
	"time"
)

// server is sent a form or returns JSON data

func handlerAPIQuotesSubmit(w http.ResponseWriter, r *http.Request) {
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
	//TODO handle error
	database.CreateQuote(database.QuoteT{
		Text:      r.FormValue("text"),
		Context:   r.FormValue("context"),
		TeacherID: uint32(teacherid),
		Unixtime:  uint64(time.Now().Unix())})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
