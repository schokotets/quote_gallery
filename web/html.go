package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"quote_gallery/database"
	"sort"
	"strconv"
	"strings"
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

func pageAdminTeachersIDEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happen as handlerAPIUnverifiedQuotes is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		return
	}

	teacher, err := database.GetTeacherByID(int32(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get teacher #%v: %v", id, err)
		return
	}

	tmpl := template.Must(template.ParseFiles("pages/edit-teacher.html"))
	tmpl.Execute(w, teacher)
}

func pageAdminTeachersAdd(w http.ResponseWriter, r *http.Request) {
	t := database.TeacherT{}
	// parse ?name=Title Name (Note) with Title and Note being optional
	// this is not expected to be perfect
	if queryParam, ok := r.URL.Query()["name"]; ok {
		name := queryParam[0]
		parts := strings.SplitN(name, " ", 3)

		switch len(parts) {
		case 0:
		case 1:
			t.Name = strings.Trim(parts[0], " ")
		case 2:
			t.Title = strings.Trim(parts[0], " ")
			t.Name = strings.Trim(parts[1], " ")
		default:
			t.Title = strings.Trim(parts[0], " ")
			t.Name = strings.Trim(parts[1], " ")
			noteWithoutParentheses := strings.ReplaceAll(strings.ReplaceAll(parts[2], "(", ""), ")", "")
			t.Note = strings.Trim(noteWithoutParentheses, " ")
		}
	}
	tmpl := template.Must(template.ParseFiles("pages/add-teacher.html"))
	tmpl.Execute(w, t)
}

func pageSubmit(w http.ResponseWriter, r *http.Request) {
	teachers, err := database.GetTeachers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get teachers")
		return
	}
	sort.Slice(teachers, func(i, j int) bool { return teachers[i].Name < teachers[j].Name })
	tmpl := template.Must(template.ParseFiles("pages/submit.html"))
	tmpl.Execute(w, teachers)
}
