// Create docker instance with postgres databasemanager inside
// sudo docker run --name some-postgres -e POSTGRES_PASSWORD=1234 -e POSTGRES_DB=quote_gallery -p 0.0.0.0:5432:5432 -d postgres

package database

import (
	"database/sql"
	"errors"
	"log"

	// loading postgresql driver
	_ "github.com/lib/pq"
)

/* -------------------------------------------------------------------------- */
/*                                Public Types                                */
/* -------------------------------------------------------------------------- */

// Teacher struct
type Teacher struct {
	ID    int
	Name  string
	Title string
	Note  string
}

// Quote struct
type Quote struct {
	ID          int
	Teacher     Teacher
	Text        string
	MatchRating float32
}

type searchCacheType struct {
	totalCount    int
	countByQuotes map[int]int
}

var database *sql.DB
var searchCache map[string]searchCacheType

/* -------------------------------------------------------------------------- */
/*                              Public Functions                              */
/* -------------------------------------------------------------------------- */

// SetupDatabase is a function to setup the database
func SetupDatabase() {
	var err error

	database, err = sql.Open("postgres", "user=postgres password=1234 dbname=quote_gallery sslmode=disable")
	if err != nil {
		log.Fatal("From SetupDatabase: ", err)
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS teachers (id serial PRIMARY KEY, name varchar, note varchar, title varchar)")
	if err != nil {
		database.Close()
		log.Fatal("From SetupDatabase: ", err)
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS quotes (id serial PRIMARY KEY, teacherid integer REFERENCES teachers (id), text varchar)")
	if err != nil {
		database.Close()
		log.Fatal("From SetupDatabase: ", err)
	}

	// Setup Search-Cache
	setupSearchCache()
}

// StoreQuote is a function to store quotes
func StoreQuote(text string, teacherid int) error {
	database.Ping()

	_, err := database.Exec("INSERT INTO quotes (teacherid, text) VALUES ($1, $2)", teacherid, text)
	if err != nil {
		log.Print("From StoreQuote: ", err)
		return err
	}

	return nil
}

// StoreTeacher is a function to store teachers
func StoreTeacher(name string, title string, note string) error {
	database.Ping()

	_, err := database.Exec("INSERT INTO teachers (name, title, note) VALUES ($1, $2, $3)", name, title, note)
	if err != nil {
		log.Print("From StoreTeacher: ", err)
		return err
	}

	return nil
}

// GetTeachers is a function to get teachers from the database
func GetTeachers() ([]Teacher, error) {

	rows, err := database.Query("SELECT id,name,note,title FROM teachers")
	defer rows.Close()

	if err != nil {
		log.Print("From GetTeachers: ", err)
		return nil, err
	}

	var teachers []Teacher

	for rows.Next() {
		t := Teacher{}
		rows.Scan(&t.ID, &t.Name, &t.Note, &t.Title)

		teachers = append(teachers, t)
	}

	return teachers, nil
}

// GetQuotes is a function to get quotes from the database
func GetQuotes() ([]Quote, error) {

	rows, err := database.Query("SELECT quotes.id, quotes.text, teachers.id, teachers.name, teachers.title, teachers.note FROM quotes INNER JOIN teachers ON quotes.teacherid = teachers.id")
	defer rows.Close()

	if err != nil {
		log.Print("From GetQuotes: ", err)
		return nil, err
	}

	var quotes []Quote

	for rows.Next() {
		q := Quote{}
		q.MatchRating = 0
		rows.Scan(&q.ID, &q.Text, &q.Teacher.ID, &q.Teacher.Name, &q.Teacher.Title, &q.Teacher.Note)
		quotes = append(quotes, q)
	}

	return quotes, nil
}

// CloseDatabase is a function to close the database
func CloseDatabase() {
	database.Close()
}

/* -------------------------------------------------------------------------- */
/*                              Private Functions                             */
/* -------------------------------------------------------------------------- */

func setupSearchCache() error {

	log.Print("Creating Search-Cache from database")

	// initialize character lookup table for words()
	setupCharacterLookup()

	// initialize the Search-Cache
	searchCache = make(map[string]searchCacheType)

	// Get all quotes from database
	rows, err := database.Query("SELECT quotes.id, quotes.text FROM quotes")
	defer rows.Close()

	if err != nil {
		log.Print("From setupSearchCache: ", err)
		return err
	}

	//Iterrate over all quotes from database
	for rows.Next() {
		// Get id and text of quote
		var text string = ""
		var id int = 0
		rows.Scan(&id, &text)

		// Iterrate over all words of quote
		for _, word := range words(text) {
			searchCacheElement := searchCache[word]
			searchCacheElement.totalCount++

			if searchCacheElement.countByQuotes == nil {
				searchCacheElement.countByQuotes = make(map[int]int)
			}

			// Read count, increment, write back
			count := searchCacheElement.countByQuotes[id]
			count++
			searchCacheElement.countByQuotes[id] = count

			searchCache[word] = searchCacheElement
		}
	}

	log.Print("Search-Cache created")

	return nil
}

func addToSearchCache(text string, id int) error {
	if searchCache == nil {
		log.Print("From addToSearchCache: searchCache is not initialized")
		return errors.New("From addToSearchCache: searchCache is not initialized")
	}

	// Iterrate over all words of quote
	for _, word := range words(text) {
		searchCacheElement := searchCache[word]
		searchCacheElement.totalCount++

		if searchCacheElement.countByQuotes == nil {
			searchCacheElement.countByQuotes = make(map[int]int)
		}

		// Read count, increment, write back
		count := searchCacheElement.countByQuotes[id]
		count++
		searchCacheElement.countByQuotes[id] = count

		searchCache[word] = searchCacheElement
	}

	return nil
}
