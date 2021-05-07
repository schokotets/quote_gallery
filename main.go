package main

import (
	"log"
	"quote_gallery/database"
	"quote_gallery/web"
)

func main() {
	log.Print("Connecting to database")
	database.Connect()
	defer database.CloseAndClearCache()

	err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Starting website on :8080")
	web.SetupRoutes()
	web.StartWebserver()
}
