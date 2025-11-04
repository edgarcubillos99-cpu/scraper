package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/eddgaroso/go-colly-mysql/internal/db"
	"github.com/eddgaroso/go-colly-mysql/internal/scraper"
	"github.com/joho/godotenv"
)

func main() {
	// cargar .env si existe
	_ = godotenv.Load()

	d, err := db.Connect()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	opts := scraper.ScrapeOptions{
		URL:           os.Getenv("TARGET_URL"),     // pon la URL en .env
		TableSelector: os.Getenv("TABLE_SELECTOR"), // ejemplo: "table#data"
		RowSelector:   os.Getenv("ROW_SELECTOR"),   // ejemplo: "tr"
		StartAtHeader: true,
	}

	if err := scraper.RunScrape(ctx, d, opts); err != nil {
		log.Println("Scrape error:", err)
	} else {
		log.Println("Scrape finalizado correctamente")
	}
}
