package web

import (
	"encoding/json"
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
		goto badRequest
	}

	// Check validity of temporary QuoteSubmission and
	// copy content into UnverifiedQuote
	switch subm.Teacher.(type) {
	case int:
		if subm.Teacher.(uint32) == 0 {
			goto badRequest
		}
		quote.TeacherID = subm.Teacher.(uint32)
		quote.TeacherName = ""
	case string:
		quote.TeacherID = 0
		quote.TeacherName = subm.Teacher.(string)
	default:
		goto badRequest
	}

	if len(subm.Context) == 0 || len(subm.Text) == 0 {
		goto badRequest
	} else {
		quote.Context = subm.Context
		quote.Text = subm.Text
	}

	// Add further information to UnverifiedQuote
	quote.Unixtime = uint64(time.Now().Unix())
	quote.IPHash = hash(strings.Split(r.RemoteAddr, ":")[0])

	// Store UnverifiedQuote in database
	database.CreateUnverifiedQuote(quote)
	return

badRequest:
	w.WriteHeader(http.StatusBadRequest)
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
