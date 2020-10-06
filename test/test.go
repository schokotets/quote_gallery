package main

import (
	"log"
	"quote_gallery/database"
)

func main() {
	log.Print("Connecting to database on :5432")
	database.Setup()
	defer database.CloseAndClearCache()

	err := database.ExecuteQuery("START TRANSACTION")
	if err != nil {
		log.Fatal("Cannot start transaction: ", err)
	}
	err = database.ExecuteQuery("DELETE FROM unverifiedQuotes")
	if err != nil {
		log.Fatal("Cannot temp-delete quotes: ", err)
	}
	err = database.ExecuteQuery("DELETE FROM quotes")
	if err != nil {
		log.Fatal("Cannot temp-delete quotes: ", err)
	}
	err = database.ExecuteQuery("DELETE FROM teachers")
	if err != nil {
		log.Fatal("Cannot temp-delete teachers: ", err)
	}

	// database.CreateTeacher(database.TeacherT{Name: "Heimburg", Title: "Herr", Note: "Sp Ge"})
	// database.CreateTeacher(database.TeacherT{Name: "Spreer", Title: "Frau", Note: "Sp Eth"})
	// database.CreateTeacher(database.TeacherT{Name: "Eidner", Title: "Frau", Note: "Sp Eth"})
	// database.CreateTeacher(database.TeacherT{Name: "Krug", Title: "Herr", Note: "Ma Ph"})
	//database.CreateTeacher("Krug", "Herr", "")
	//database.CreateTeacher("Spreer", "Frau", "")

	// i := database.GetTeachers()
	// fmt.Println(i)

	// database.CreateQuote(database.QuoteT{
	// 	TeacherID: (*i)[0].TeacherID,
	// 	Context:   "Nicer Tag",
	// 	Text:      "AAA BBB CCC",
	// })

	// database.CreateQuote(database.QuoteT{
	// 	TeacherID: (*i)[0].TeacherID,
	// 	Context:   "Nicer Tag",
	// 	Text:      "BBB CCC",
	// })

	// database.CreateQuote(database.QuoteT{
	// 	TeacherID: (*i)[0].TeacherID,
	// 	Context:   "Nicer Tag",
	// 	Text:      "DDD EEE",
	// })

	// j := database.GetQuotes()
	// fmt.Println(j)

	// database.UpdateQuote(database.QuoteT{
	// 	QuoteID:   2,
	// 	TeacherID: (*i)[0].TeacherID,
	// 	Context:   "Nicer Tag",
	// 	Text:      "BBB CCC DDD",
	// })

	// j = database.GetQuotes()
	// fmt.Println(j)

	// j = database.GetQuotesByString("BBB DDD")
	// fmt.Println(j)

	log.Print(database.CreateUnverifiedQuote(database.UnverifiedQuoteT{
		QuoteID:     0,
		TeacherID:   0,
		TeacherName: "Test",
		Context:     "sadklfjh",
		Text:        "sadlkfusdl",
		IPHash:      3487562938475,
		Unixtime:    23489576485,
	}))

	//j, _ := database.GetQuotes()
	//fmt.Println(j)

	database.ExecuteQuery("ROLLBACK")
}
