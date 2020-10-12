package main

import (
	"log"
	"quote_gallery/database"
	"quote_gallery/web"
)

func main() {
	log.Print("Connecting to database on :5432")
	database.Connect()
	defer database.CloseAndClearCache()

	database.Initialize()

	log.Print("Starting website on :8080")
	web.SetupRoutes()
	web.StartWebserver()
}
