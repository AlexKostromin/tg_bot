package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/AlexKostromin/tg_bot/internal/config"
	"github.com/AlexKostromin/tg_bot/internal/db"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: createadmin <username> <password>\n")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]

	godotenv.Load()
	cfg := config.Load()

	pg, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB connection error: %v\n", err)
		os.Exit(1)
	}
	defer pg.Close()

	if err := db.RunMigration(pg); err != nil {
		fmt.Fprintf(os.Stderr, "Migration error: %v\n", err)
		os.Exit(1)
	}

	repo := repository.NewAdminUserRepository(pg)

	user, err := repo.Create(context.Background(), username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Admin user created: id=%d, username=%s\n", user.ID, user.Username)
}
