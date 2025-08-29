package main

import (
	"log"

	"awesomeProject1/internal/config"
	"awesomeProject1/internal/database"
)

func main() {
	cfg := config.LoadConfig()
	log.Printf("Running migrations on database: %s\n", cfg.DatabasePath)

	if err := database.RunMigrations(cfg.DatabasePath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
}
