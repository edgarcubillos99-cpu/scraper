package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"

	"github.com/eddgaroso/go-colly-mysql/internal/db"
	"github.com/eddgaroso/go-colly-mysql/internal/scraper"
)

func runScraper() {
	// Conectar DB
	database, err := db.Connect()
	if err != nil {
		log.Printf("❌ Error conectando DB: %v", err)
		return
	}

	// Collector + cookies
	c, jar, err := scraper.InitCollector()
	if err != nil {
		log.Printf("❌ Error Collector: %v", err)
		return
	}

	// Login
	loginInfo := scraper.LoginInfo{
		Username: os.Getenv("BILLING_USER"),
		Password: os.Getenv("BILLING_PASS"),
	}

	if err := scraper.Login(c, jar, loginInfo); err != nil {
		log.Printf("❌ Error login: %v", err)
		return
	}

	// Calcular fechas
	loc, _ := time.LoadLocation(os.Getenv("TIMEZONE"))

	startOfDay := time.Now().In(loc).Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)

	urlTemplate := os.Getenv("TARGET_URL")
	targetURL := fmt.Sprintf(urlTemplate, startOfDay.Unix(), endOfDay.Unix())

	opts := scraper.ScrapeOptions{
		URL:           targetURL,
		TableSelector: os.Getenv("TABLE_SELECTOR"),
		RowSelector:   os.Getenv("ROW_SELECTOR"),
		StartAtHeader: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	// Ejecutar scraping
	log.Println("🚀 Ejecutando scraping programado...")
	if err := scraper.RunScrape(ctx, database, c, opts); err != nil {
		log.Printf("❌ Error scraping: %v", err)
	}
}

func main() {

	godotenv.Load()

	// Leer hora desde variables de entorno
	hour := os.Getenv("SCRAPER_HOUR")
	min := os.Getenv("SCRAPER_MINUTE")
	tz := os.Getenv("TIMEZONE")

	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Fatalf("❌ TIMEZONE inválido: %v", err)
	}

	s := gocron.NewScheduler(loc)

	// Programar tarea diaria
	_, err = s.Every(1).Day().At(fmt.Sprintf("%s:%s", hour, min)).Do(runScraper)
	if err != nil {
		log.Fatalf("❌ Error programando tarea: %v", err)
	}

	log.Printf("✅ Scraper programado cada día a las %s:%s (%s)\n", hour, min, tz)

	s.StartBlocking() // Mantiene el container vivo
}
