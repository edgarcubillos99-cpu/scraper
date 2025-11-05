package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eddgaroso/go-colly-mysql/internal/db"      // módulo de conexión a DB
	"github.com/eddgaroso/go-colly-mysql/internal/scraper" // módulo de scraper
	"github.com/joho/godotenv"                             // para cargar variables de entorno
)

// principal función del scraper
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

	// cargar zona horaria desde variable de entorno
	loc, _ := time.LoadLocation(os.Getenv("America/Puerto_Rico"))

	// definir rango de fechas del día actual en la zona horaria especificada
	startOfDay := time.Now().In(loc).Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)

	// construir URL con timestamps
	urlTemplate := os.Getenv("TARGET_URL") // pon la URL en yml
	url := fmt.Sprintf(urlTemplate, startOfDay.Unix(), endOfDay.Unix())

	opts := scraper.ScrapeOptions{
		URL:           url,                         // URL con timestamps
		TableSelector: os.Getenv("TABLE_SELECTOR"), // ejemplo: "table#data"
		RowSelector:   os.Getenv("ROW_SELECTOR"),   // ejemplo: "tr"
		StartAtHeader: true,                        // ajusta según la tabla
	}

	// ejecutar el scraping
	if err := scraper.RunScrape(ctx, d, opts); err != nil {
		log.Println("Scrape error:", err)
	} else {
		log.Println("Scrape finalizado correctamente")
	}
}
