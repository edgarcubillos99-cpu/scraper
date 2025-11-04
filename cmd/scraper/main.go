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
	_ = godotenv.Load() // godotenv no falla si no existe el archivo

	// conectar a la base de datos
	d, err := db.Connect()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	// contexto con timeout para la operación de scraping
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// opciones de scraping
	opts := scraper.ScrapeOptions{
		URL:           os.Getenv("TARGET_URL"),     // pon la URL en .env
		TableSelector: os.Getenv("TABLE_SELECTOR"), // ejemplo: "table#data"
		RowSelector:   os.Getenv("ROW_SELECTOR"),   // ejemplo: "tr"
		StartAtHeader: true,
	}

	// ejecutar el scraping
	if err := scraper.RunScrape(ctx, d, opts); err != nil {
		log.Println("Scrape error:", err)
	} else {
		log.Println("Scrape finalizado correctamente")
	}
}
