package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/eddgaroso/go-colly-mysql/internal/db"
	"github.com/eddgaroso/go-colly-mysql/internal/scraper"
)

func main() {

	// ------------------------------------------------------
	// Cargar variables de entorno (.env es opcional)
	// ------------------------------------------------------
	_ = godotenv.Load()

	// ------------------------------------------------------
	// Conectar a la base de datos
	// ------------------------------------------------------
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("❌ Error conectando a DB: %v", err)
	}

	log.Println("✅ Conectado a la base de datos.")

	// ------------------------------------------------------
	// Crear collector + jar para cookies
	// ------------------------------------------------------
	c, jar, err := scraper.InitCollector()
	if err != nil {
		log.Fatalf("❌ Error iniciando Collector: %v", err)
	}

	// ------------------------------------------------------
	// Hacer login
	// ------------------------------------------------------
	loginInfo := scraper.LoginInfo{
		Username: os.Getenv("BILLING_USER"),
		Password: os.Getenv("BILLING_PASS"),
	}

	log.Println("🔐 Realizando login...")

	if err := scraper.Login(c, jar, loginInfo); err != nil {
		log.Fatalf("❌ Error en login: %v", err)
	}

	log.Println("✅ Login exitoso. Cookies activas.")

	// ------------------------------------------------------
	// Construir URL con timestamps
	// ------------------------------------------------------
	loc, _ := time.LoadLocation(os.Getenv("TIMEZONE"))

	startOfDay := time.Now().In(loc).Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)

	urlTemplate := os.Getenv("TARGET_URL")
	targetURL := fmt.Sprintf(urlTemplate, startOfDay.Unix(), endOfDay.Unix())

	// ------------------------------------------------------
	// Configurar opciones del scraper
	// ------------------------------------------------------
	opts := scraper.ScrapeOptions{
		URL:           targetURL,
		TableSelector: os.Getenv("TABLE_SELECTOR"),
		RowSelector:   os.Getenv("ROW_SELECTOR"),
		StartAtHeader: false,
	}

	// ------------------------------------------------------
	// Ejecutar scraping
	// ------------------------------------------------------
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	log.Println("🔎 Iniciando scraping...")

	if err := scraper.RunScrape(ctx, database, c, opts); err != nil {
		log.Fatalf("❌ Error durante scraping: %v", err)
	}

	log.Println("✅ Scraping finalizado sin errores.")
}
