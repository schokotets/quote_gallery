package web

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"quote_gallery/database"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

/* -------------------------------------------------------------------------- */
/*                                 DEFINITIONS                                */
/* -------------------------------------------------------------------------- */

type quoteInputT struct {
	Teacher interface{}
	Context string
	Text    string
}

type teacherInputT struct {
	Name  string
	Title string
	Note  string
}

/* -------------------------------------------------------------------------- */
/*                           EXPORTED API FUNCTIONS                           */
/* -------------------------------------------------------------------------- */

func handlerAPIQuotesSubmit(w http.ResponseWriter, r *http.Request) {
	var subm quoteInputT
	var quote database.UnverifiedQuoteT

	// Check if right http method is used
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// parse json request body into temporary QuoteInput
	bytes, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(bytes, &subm)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unparsable JSON")
		return
	}

	// Check validity of temporary QuoteInput and
	// copy content into UnverifiedQuote

	switch subm.Teacher.(type) {
	case float64:
		if subm.Teacher.(float64) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "invalid TeacherID: 0")
			return
		}
		quote.TeacherID = int32(subm.Teacher.(float64))
		quote.TeacherName = ""
	case string:
		quote.TeacherID = 0
		quote.TeacherName = subm.Teacher.(string)
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid TeacherID: its type is neither string nor int")
		return
	}

	if len(subm.Text) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Text is empty")
		return
	}

	quote.Context = subm.Context
	quote.Text = subm.Text

	// Add further information to UnverifiedQuote
	quote.Unixtime = int64(time.Now().Unix())
	quote.IPHash = hash(strings.Split(r.RemoteAddr, ":")[0])

	// Store UnverifiedQuote in database
	status := database.CreateUnverifiedQuote(quote)

	switch status.Code {
	case database.StatusError:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		log.Printf("/api/quotes/submit: quote creation failed with error '%v' for request body '%s' and UnverifiedQuoteT %v", status.Message, bytes, quote)
		break
	case database.StatusInvalidTeacherID:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Teacher: no teacher with that ID")
		break
	}

	return
}

func handlerAPIUnverifiedQuotesID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happend, because handlerAPIUnverifiedQuotes is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		return
	}
	if id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid QuoteID: 0")
		return
	}

	var subm quoteInputT
	var quote database.UnverifiedQuoteT

	switch r.Method {
	case "PUT":

		// parse json request body into temporary QuoteInput
		bytes, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(bytes, &subm)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "unparsable JSON")
			return
		}

		// Check validity of temporary QuoteInput and
		// copy content into UnverifiedQuote

		switch subm.Teacher.(type) {
		case float64:
			if subm.Teacher.(float64) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "invalid TeacherID: 0")
				return
			}
			quote.TeacherID = int32(subm.Teacher.(float64))
			quote.TeacherName = ""
		case string:
			quote.TeacherID = 0
			quote.TeacherName = subm.Teacher.(string)
		default:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "invalid TeacherID: its type is neither string nor int")
			return
		}

		if len(subm.Text) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Text is empty")
			return
		}

		quote.QuoteID = int32(id)
		quote.Context = subm.Context
		quote.Text = subm.Text

		// Update UnverifiedQuote in database
		status := database.UpdateUnverifiedQuote(quote)

		switch status.Code {
		case database.StatusError:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/unverifiedquotes/:id: quote updating failed with error '%v' for request body '%s' and UnverifiedQuoteT %v", status.Message, bytes, quote)
			break
		case database.StatusInvalidTeacherID:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown TeacherID: %d", quote.TeacherID)
			break
		case database.StatusInvalidQuoteID:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown QuoteID: %d", quote.QuoteID)
			break
		}

		break

	case "DELETE":

		// Delete UnverifiedQuote from database
		status := database.DeleteUnverifiedQuote(int32(id))

		switch status.Code {
		case database.StatusError:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/unverifiedquotes/:id: quote deletion failed with error '%v'", status.Message)
			break
		case database.StatusInvalidQuoteID:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown QuoteID: %d", id)
			break
		}

		break

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		break
	}
}

func handlerAPIUnverifiedQuotesIDConfirm(w http.ResponseWriter, r *http.Request) {
	// Check if right http method is used
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happend, because handlerAPIUnverifiedQuotes is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		return
	}

	if id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid QuoteID: 0")
		return
	}

	q, status := database.GetUnverifiedQuoteByID(int32(id))

	switch status.Code {
	case database.StatusError:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		log.Printf("/api/unverifiedquotes/:id/confirm: getting unverified quotes failed with error '%v'", status.Message)
		return
	case database.StatusInvalidQuoteID:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "unknown QuoteID: %d", id)
		return
	}

	if q.TeacherID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid TeacherID: 0")
		return
	}

	status = database.CreateQuote(database.QuoteT{
		TeacherID: q.TeacherID,
		Context:   q.Context,
		Text:      q.Text,
		Unixtime:  q.Unixtime,
	})

	switch status.Code {
	case database.StatusError:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		log.Printf("/api/unverifiedquotes/:id/confirm: quote creation failed with error '%v'", status.Message)
		return
	case database.StatusInvalidTeacherID:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "unknown TeacherID: %d", id)
		return
	}

	database.DeleteUnverifiedQuote(int32(id))
}

func handlerAPITeachers(w http.ResponseWriter, r *http.Request) {
	// Check if right http method is used
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var subm teacherInputT
	var teacher database.TeacherT

	// parse json request body into temporary QuoteInput
	bytes, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(bytes, &subm)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unparsable JSON")
		return
	}

	if len(subm.Name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Name is empty")
		return
	}

	if len(subm.Title) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Title is empty")
		return
	}

	teacher.Name = subm.Name
	teacher.Title = subm.Title
	teacher.Note = subm.Note

	status := database.CreateTeacher(teacher)

	if status.Code != database.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		log.Printf("/api/teachers: creating teacher failed with error '%v' for request body '%s' and TeacherT %v", status.Message, bytes, teacher)
		return
	}
}

func handlerAPITeachersID(w http.ResponseWriter, r *http.Request) {
	// Check if right http method is used
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happend, because handlerAPIUnverifiedQuotes is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		return
	}

	if id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid TeacherID: 0")
		return
	}

	var subm teacherInputT
	var teacher database.TeacherT

	// parse json request body into temporary QuoteInput
	bytes, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(bytes, &subm)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unparsable JSON")
		return
	}

	if len(subm.Name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Name is empty")
		return
	}

	if len(subm.Title) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Title is empty")
		return
	}

	teacher.TeacherID = int32(id)
	teacher.Name = subm.Name
	teacher.Title = subm.Title
	teacher.Note = subm.Note

	status := database.UpdateTeacher(teacher)

	if status.Code != database.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		log.Printf("/api/teachers: updating teacher failed with error '%v' for request body '%s' and TeacherT %v", status.Message, bytes, teacher)
		return
	}
}

/* -------------------------------------------------------------------------- */
/*                         UNEXPORTED HELPER FUNCTIONS                        */
/* -------------------------------------------------------------------------- */

func hash(s string) int64 {
	x := fnv.New64a()
	x.Write([]byte(s))
	return int64(x.Sum64())
}
