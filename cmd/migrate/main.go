package main

import (
	"log"
	"os"

	"awesomeProject1/internal/database"
)

func main() {
	// Get database path from command line argument or use default
	dbPath := "./app.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	log.Printf("Running migrations on database: %s\n", dbPath)

	if err := database.RunMigrations(dbPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
}
