package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
)

func main() {
	_ = godotenv.Load()

	if err := repository.RunMigrations(context.Background()); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migrations applied")
}
