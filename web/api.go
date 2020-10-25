package web

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
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
	err = database.CreateUnverifiedQuote(quote)

	/* ------------------------- MISSING ERROR HANDLING ------------------------- */
	// if err != nil {
	// 	if strings.Contains(err.Error(), `violates foreign key constraint`) {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		fmt.Fprintf(w, "Teacher: no teacher with that ID")
	// 	} else {
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		fmt.Fprintf(w, "internal server error")
	// 		log.Printf("/api/quotes/submit: quote creation failed with error '%v' for request body '%s' and UnverifiedQuoteT %v", err, bytes, quote)
	// 	}
	// }
}

func handlerAPIUnverifiedQuotes(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprintf(w, "invalid UnverifiedQuoteID: 0")
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

		quote.Context = subm.Context
		quote.Text = subm.Text

		// Update UnverifiedQuote in database
		err = database.UpdateUnverifiedQuote(quote)

		/* ------------------------- MISSING ERROR HANDLING ------------------------- */
		// if err != nil {
		// 	if strings.Contains(err.Error(), `violates foreign key constraint`) {
		// 		w.WriteHeader(http.StatusBadRequest)
		// 		fmt.Fprintf(w, "Teacher: no teacher with that ID")
		// 	} else {
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		fmt.Fprintf(w, "internal server error")
		// 		log.Printf("/api/unverifiedquotes/:id: quote creation failed with error '%v' for request body '%s' and UnverifiedQuoteT %v", err, bytes, quote)
		// 	}
		// }
		break

	case "DELETE":

		// Delete UnverifiedQuote from database
		err = database.DeleteUnverifiedQuote(int32(id))

		/* ------------------------- MISSING ERROR HANDLING ------------------------- */

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func handlerAPIUnverifiedQuotesConfirm(w http.ResponseWriter, r *http.Request) {

}

/* -------------------------------------------------------------------------- */
/*                         UNEXPORTED HELPER FUNCTIONS                        */
/* -------------------------------------------------------------------------- */

func hash(s string) int64 {
	x := fnv.New64a()
	x.Write([]byte(s))
	return int64(x.Sum64())
}
