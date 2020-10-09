// Create docker instance with postgres databasemanager inside
// sudo docker run --name some-postgres -e POSTGRES_PASSWORD=1234 -e POSTGRES_DB=quote_gallery -p 0.0.0.0:5432:5432 -d postgres

package database

import (
	"database/sql"
	"log"

	// loading postgresql driver
	_ "github.com/lib/pq"
)

// Teacher struct
type Teacher struct {
	ID    int
	Name  string
	Title string
	Note  string
}

// Quote struct
type Quote struct {
	ID      int
	Teacher Teacher
	Text    string
}

var database *sql.DB

// SetupDatabase is a function to setup the database
func SetupDatabase() {
	var err error

	database, err = sql.Open("postgres", "user=postgres password=1234 dbname=quote_gallery sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open Database: ", err)
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS teachers (id serial PRIMARY KEY, name varchar, note varchar, title varchar)")
	if err != nil {
		database.Close()
		log.Fatal("Cannot create teachers table: ", err)
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS quotes (id serial PRIMARY KEY, teacherid integer REFERENCES teachers (id), text varchar)")
	if err != nil {
		database.Close()
		log.Fatal("Cannot create quotes table: ", err)
	}
}

// StoreQuote is a function to store quotes
func StoreQuote(quote string, teacherid int) error {
	database.Ping()

	_, err := database.Exec("INSERT INTO quotes (teacherid, text) VALUES ($1, $2)", teacherid, quote)
	if err != nil {
		log.Print("Cannot store quote: ", err)
		return err
	}

	return nil
}

// StoreTeacher is a function to store teachers
func StoreTeacher(name string, title string, note string) error {
	database.Ping()

	_, err := database.Exec("INSERT INTO teachers (name, title, note) VALUES ($1, $2, $3)", name, title, note)
	if err != nil {
		log.Print("Cannot store teacher: ", err)
		return err
	}

	return nil
}

// GetTeachers is a function to get teachers from the database
func GetTeachers() ([]Teacher, error) {

	rows, err := database.Query("SELECT id,name,note,title FROM teachers")
	defer rows.Close()

	if err != nil {
		log.Print("Cannot get teachers: ", err)
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
		log.Print("Cannot get quotes: ", err)
		return nil, err
	}

	var quotes []Quote

	for rows.Next() {
		q := Quote{}
		rows.Scan(&q.ID, &q.Text, &q.Teacher.ID, &q.Teacher.Name, &q.Teacher.Title, &q.Teacher.Note)
		quotes = append(quotes, q)
	}

	return quotes, nil
}

// CloseDatabase is a function to close the database
func CloseDatabase() {
	database.Close()
}
