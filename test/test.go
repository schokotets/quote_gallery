package main

import (
	"fmt"
	"log"
	"quote_gallery/database"
)

func main() {
	log.Print("Connecting to database on :5432")
	database.SetupDatabase()
	defer database.CloseDatabase()

	err := database.ExecuteQuery("START TRANSACTION")
	if err != nil {
		log.Fatal("Cannot start transaction: ", err)
	}
	err = database.ExecuteQuery("DELETE FROM quotes")
	if err != nil {
		log.Fatal("Cannot temp-delete quotes: ", err)
	}
	err = database.ExecuteQuery("DELETE FROM teachers")
	if err != nil {
		log.Fatal("Cannot temp-delete teachers: ", err)
	}

	database.StoreTeacher("Heimburg", "Herr", "")
	database.StoreTeacher("Krug", "Herr", "")
	database.StoreTeacher("Spreer", "Frau", "")

	i, _ := database.GetTeachers()
	fmt.Println(i)

	database.StoreQuote("Ich mach dich rund, wi'n Buslenker", i[0].ID)
	database.StoreQuote("Brust steif machen und mit'm Nippl annehmen.", i[0].ID)
	database.StoreQuote("Mathe ma' dick! Mathe mal dünn!", i[1].ID)
	database.StoreQuote("Sport ist Mord und Massensport ist Massenmord.", i[1].ID)
	database.StoreQuote("Lass mein Hütchen, das ist Friedhelm!", i[2].ID)

	j, _ := database.GetQuotes()
	fmt.Println(j)

	database.ExecuteQuery("ROLLBACK")
}
