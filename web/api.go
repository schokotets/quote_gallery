package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"quote_gallery/database"
	"strconv"
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

func postAPIQuotesSubmit(w http.ResponseWriter, r *http.Request, u int32) {
	var subm quoteInputT
	var quote database.UnverifiedQuoteT

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
		if int32(subm.Teacher.(float64)) <= 0 {
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

	quote.UserID = u

	// Add further information to UnverifiedQuote
	quote.Unixtime = int64(time.Now().Unix())

	// Store UnverifiedQuote in database
	err = database.CreateUnverifiedQuote(quote)

	if err != nil {
		switch err.(type) {
		case database.InvalidTeacherIDError:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Teacher: no teacher with that ID")
		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/quotes/submit: quote creation failed with error '%s' for request body '%s' and UnverifiedQuoteT %v", err.Error(), bytes, quote)
		}
	}
}

func putAPIUnverifiedQuotesID(w http.ResponseWriter, r *http.Request, u int32) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happend, because this handler is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot convert in-url id to int")
		return
	}
	if id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid QuoteID: 0")
		return
	}

	var subm quoteInputT
	var quote database.UnverifiedQuoteT

	// parse json request body into temporary QuoteInput
	bytes, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(bytes, &subm)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unparsable JSON")
		return
	}

	// Check validity of temporary QuoteInput and
	// copy content into UnverifiedQuote

	switch subm.Teacher.(type) {
	case float64:
		if int32(subm.Teacher.(float64)) <= 0 {
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
	err = database.UpdateUnverifiedQuote(quote)

	if err != nil {
		switch err.(type) {
		case database.InvalidTeacherIDError:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown TeacherID: %d", quote.TeacherID)
		case database.InvalidQuoteIDError:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown QuoteID: %d", quote.QuoteID)
		default: //generic / database.DBError:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/unverifiedquotes/:id: quote updating failed with error '%s' for request body '%s' and UnverifiedQuoteT %v", err.Error(), bytes, quote)
		}
	}
}

func deleteAPIUnverifiedQuotesID(w http.ResponseWriter, r *http.Request, u int32) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happend, because this handler is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot convert in-url id to int")
		return
	}
	if id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid QuoteID: 0")
		return
	}

	// Delete UnverifiedQuote from database
	err = database.DeleteUnverifiedQuote(int32(id))

	if err != nil {
		switch err.(type) {
		case database.InvalidQuoteIDError:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown QuoteID: %d", id)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/unverifiedquotes/:id: quote deletion failed with error '%s'", err.Error())
		}
	}
}

func putAPIUnverifiedQuotesIDConfirm(w http.ResponseWriter, r *http.Request, u int32) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happend, because this handler is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot convert in-url id to int")
		return
	}

	if id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid QuoteID: 0")
		return
	}

	q, err := database.GetUnverifiedQuoteByID(int32(id))

	if err != nil {
		switch err.(type) {
		case database.InvalidQuoteIDError:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown QuoteID: %d", id)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/unverifiedquotes/:id/confirm: getting unverified quotes failed with error '%s'", err.Error())
			return
		}
	}

	if q.TeacherID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unverifiedQuote has invalid TeacherID (it needs a valid one to be confirmed)")
		return
	}

	err = database.CreateQuote(database.QuoteT{
		TeacherID: q.TeacherID,
		Context:   q.Context,
		Text:      q.Text,
		Unixtime:  q.Unixtime,
	})

	if err != nil {
		switch err.(type) {
		case database.InvalidTeacherIDError:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown TeacherID: %d", id)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/unverifiedquotes/:id/confirm: quote creation failed with error '%s'", err.Error())
			return
		}
	}

	database.DeleteUnverifiedQuote(int32(id))
}

func putAPIUnverifiedQuotesIDAssignTeacherID(w http.ResponseWriter, r *http.Request, u int32) {
	quoteid, err := strconv.Atoi(mux.Vars(r)["quoteid"])
	if err != nil {
		// This should not happen because this handler is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot convert in-url id to int")
		return
	}

	if quoteid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid QuoteID: 0")
		return
	}

	teacherid, err := strconv.Atoi(mux.Vars(r)["teacherid"])
	if err != nil {
		// This should not happen, see above
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot convert in-url id to int")
		return
	}

	if teacherid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid TeacherID: 0")
		return
	}

	q, err := database.GetUnverifiedQuoteByID(int32(quoteid))

	if err != nil {
		switch err.(type) {
		case database.InvalidQuoteIDError:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown QuoteID: %d", quoteid)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/unverifiedquotes/:quoteid/assignteacher/:teacherid: getting quote failed with error '%s'", err.Error())
			return
		}
	}

	q.TeacherID = int32(teacherid)
	q.TeacherName = ""

	err = database.UpdateUnverifiedQuote(q)

	if err != nil {
		switch err.(type) {
		case database.InvalidTeacherIDError:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown TeacherID: %d", teacherid)
			return
		case database.InvalidQuoteIDError:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "unknown QuoteID: %d", quoteid)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
			log.Printf("/api/unverifiedquotes/:quoteid/assignteacher/:teacherid: assigning teacher failed with error '%s'", err.Error())
			return
		}
	}
}

func postAPITeachers(w http.ResponseWriter, r *http.Request, u int32) {
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

	err = database.CreateTeacher(teacher)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		log.Printf("/api/teachers: creating teacher failed with error '%s' for request body '%s' and TeacherT %v", err.Error(), bytes, teacher)
		return
	}
}

func putAPITeachersID(w http.ResponseWriter, r *http.Request, u int32) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happend, because this handler is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot convert in-url id to int")
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

	err = database.UpdateTeacher(teacher)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		log.Printf("/api/teachers: updating teacher failed with error '%s' for request body '%s' and TeacherT %v", err.Error(), bytes, teacher)
		return
	}
}

func putAPIQuotesIDVoteRating(w http.ResponseWriter, r *http.Request, u int32) {
	quoteid, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		// This should not happend, because this handler is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot convert in-url id to int")
		return
	}

	if quoteid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid QuoteID: 0")
		return
	}

	val, err := strconv.Atoi(mux.Vars(r)["val"])
	if err != nil {
		// This should not happend, because this handler is only called if
		// uri pattern is matched, see web.go
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot convert in-url vote val to int")
		return
	}

	quote, err := database.AddVote(database.VoteT{
		QuoteID: int32(quoteid),
		UserID: u,
		Val: int8(val),
	})

	if err != nil {
		switch err.(type) {
		case database.InvalidQuoteIDError:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "unknown QuoteID: %d", quoteid)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			log.Printf("/api/quotes/:id/vote/:val: vote casting failed with error '%s' for QuoteID %d and rating value %d", err.Error(), quoteid, val)
		}
		return
	}

	b, err := json.Marshal(quote.Stats)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "marshalling to json failed")
		return
	}

	fmt.Fprintf(w, string(b))
}
