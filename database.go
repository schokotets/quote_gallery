// Create docker instance with postgres databasemanager inside
// sudo docker run --name some-postgres -e POSTGRES_PASSWORD=1234 -e POSTGRES_DB=quote_gallery -p 0.0.0.0:5432:5432 -d postgres

package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type teacher struct {
	ID    int
	Name  string
	Title string
	Note  string
}

type quote struct {
	ID      int
	Teacher teacher
	Text    string
}

var database *sql.DB

func setupDatabase() {
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

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS quotes (id serial PRIMARY KEY, teacherid integer REFERENCES teachers (id), quote varchar)")
	if err != nil {
		database.Close()
		log.Fatal("Cannot create quotes table: ", err)
	}
}

func storeQuote(quote string, teacherid int) error {
	database.Ping()

	_, err := database.Exec("INSERT INTO quotes (teacherid, quote) VALUES ($1, $2)", teacherid, quote)
	if err != nil {
		log.Print("Cannot store quote: ", err)
		return err
	}

	return nil
}

func storeTeacher(name string, title string, note string) error {
	database.Ping()

	_, err := database.Exec("INSERT INTO teachers (name, title, note) VALUES ($1, $2, $3)", name, title, note)
	if err != nil {
		log.Print("Cannot store teacher: ", err)
		return err
	}

	return nil
}

func getTeachers() ([]teacher, error) {

	rows, err := database.Query("SELECT id,name,note,title FROM teachers")
	defer rows.Close()

	if err != nil {
		log.Print("Cannot get teachers: ", err)
		return nil, err
	}

	var teachers []teacher

	for rows.Next() {
		t := teacher{}
		rows.Scan(&t.ID, &t.Name, &t.Note, &t.Title)

		teachers = append(teachers, t)
	}

	return teachers, nil
}

func closeDatabase() {
	database.Close()
}
