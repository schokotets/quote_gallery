package web

import (
	"encoding/json"
	"fmt"
	"log"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"quote_gallery/database"
	"strings"
	"time"
)

/* -------------------------------------------------------------------------- */
/*                                 DEFINITIONS                                */
/* -------------------------------------------------------------------------- */

type quoteSubmissionT struct {
	Teacher interface{}
	Context string
	Text    string
}

/* -------------------------------------------------------------------------- */
/*                           EXPORTED API FUNCTIONS                           */
/* -------------------------------------------------------------------------- */

func handlerAPIQuotesSubmit(w http.ResponseWriter, r *http.Request) {
	var subm quoteSubmissionT
	var quote database.UnverifiedQuoteT

	// Check if right http method is used
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// parse json request body into temporary QuoteSubmission
	bytes, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(bytes, &subm)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unparsable JSON")
		return
	}

	// Check validity of temporary QuoteSubmission and
	// copy content into UnverifiedQuote

	switch subm.Teacher.(type) {
	case float64:
		if subm.Teacher.(float64) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "invalid TeacherID: 0")
			return
		}
		quote.TeacherID = uint32(subm.Teacher.(float64))
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
	quote.Unixtime = uint64(time.Now().Unix())
	quote.IPHash = hash(strings.Split(r.RemoteAddr, ":")[0])

	// Store UnverifiedQuote in database
	err = database.CreateUnverifiedQuote(quote)
	if err != nil {
		//TODO do not return InternalServerError if TeacherID invalid
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		log.Printf("/api/quotes/submit: quote creation failed with error %v for request body %s and UnverifiedQuoteT %v", err, bytes, quote)
	}

	return
}

/* -------------------------------------------------------------------------- */
/*                         UNEXPORTED HELPER FUNCTIONS                        */
/* -------------------------------------------------------------------------- */

func hash(s string) uint64 {
	x := fnv.New64a()
	x.Write([]byte(s))
	return x.Sum64()
}
