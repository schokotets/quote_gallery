package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"quote_gallery/database"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

//server returns HTML data

const quotesPerPage = 15

func pageRoot(w http.ResponseWriter, r *http.Request, userID int32, isAdmin bool) {
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

	var indexHandler *database.IndexHandler
	var indexHandlerKey string
	if sortQuery, ok := r.URL.Query()["sorting"]; ok {
		sorting := sortQuery[0]
		if ih, ok := database.IndexHandlers[sorting]; ok {
			indexHandler = &ih
			indexHandlerKey = sorting
		}
	}

	if indexHandler == nil {
		indexHandlerKey = database.DefaultIndexHandlerName
		ih, _ := database.IndexHandlers[indexHandlerKey]
		indexHandler = &ih
	}


	quotes, err := database.GetNSortedQuotesFrom(quotesPerPage, currentPage*quotesPerPage, indexHandler.Function)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if quotes == nil {
		quotes, err = database.GetNSortedQuotesFrom(quotesPerPage, 0, indexHandler.Function)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = database.AddUserDataToQuotes(quotes, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	data := struct {
		Quotes	[]database.QuoteT
		Prev	int
		Current	int
		Next	int
		Last	int
		IsAdmin bool
		SortingOrder [6]string
		SortingMap map[string]database.IndexHandler
		CurrentSorting string
	}{quotes, previousPage, currentPage, nextPage, lastPage, isAdmin, database.IndexHandlerOrder, database.IndexHandlers, indexHandlerKey}

	tmpl := template.Must(template.New("quotes.html").Funcs(template.FuncMap{
		"inc": func (i int) int { return i+1 },
		"div": func (a, b int32) string { return fmt.Sprintf("%.3f", float32(a)/float32(b)) },
		"GetTeacherByID": database.GetTeacherByID,
	}).ParseFiles("pages/quotes.html"))
	tmpl.Execute(w, data)
}

func pageAdmin(w http.ResponseWriter, r *http.Request, u int32) {
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

	sortedteachers := make([]database.TeacherT, len(teachers))
	copy(sortedteachers, teachers)
	sort.Slice(sortedteachers, func(i, j int) bool { return sortedteachers[i].Name < sortedteachers[j].Name })


	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Unixtime < quotes[j].Unixtime })
	sort.Slice(teachers, func(i, j int) bool { return teachers[i].TeacherID < teachers[j].TeacherID })

	_, showusers := r.URL.Query()["showusers"]


	pagedata := struct {
		Quotes []database.UnverifiedQuoteT
		Teachers []database.TeacherT
		SortedTeachers []database.TeacherT
		ShowUsers bool
	} {
		quotes,
		teachers,
		sortedteachers,
		showusers,
	}

	tmpl := template.Must(template.New("admin.html").Funcs(template.FuncMap{
		"GetTeacherByID": database.GetTeacherByID,
		"GetUsernameByID": database.GetUsernameByID,
		"FormatUnixtime": func(utime int64) string {
			return time.Unix(utime, 0).Format("2.1.2006 15:04")
		},
	}).ParseFiles("pages/admin.html"))

	err = tmpl.Execute(w, pagedata)
	if err != nil {
		panic(err)
	}
}

func pageAdminUnverifiedQuotesIDEdit(w http.ResponseWriter, r *http.Request, u int32) {
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

func pageAdminTeachersIDEdit(w http.ResponseWriter, r *http.Request, u int32) {
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

func pageAdminTeachersAdd(w http.ResponseWriter, r *http.Request, u int32) {
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

func pageSubmit(w http.ResponseWriter, r *http.Request, u int32) {
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

func pageSimilarQuotes(w http.ResponseWriter, r *http.Request, u int32) {
	if queryParam, ok := r.URL.Query()["text"]; ok {
		text := queryParam[0]
		if text == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "no text given")
			return
		}

		similarquotes, err := database.GetMaxNQuotesByString(3, text)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "failed to retrieve matching quotes")
			log.Printf("/suggestions: retrieving matching quotes failed with error '%s'", err.Error())
			return
		}

		if len(similarquotes) == 0 {
			fmt.Fprintf(w, "")
			return
		}

		tmpl := template.Must(template.New("suggestions.html").Funcs(template.FuncMap{
			"GetTeacherByID": database.GetTeacherByID,
		}).ParseFiles("pages/suggestions.html"))
		tmpl.Execute(w, similarquotes)
	}
}
