package scraper

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/eddgaroso/go-colly-mysql/internal/model"
	"github.com/gocolly/colly/v2"
	"gorm.io/gorm"
)

// Opciones de scraping
type ScrapeOptions struct {
	URL           string
	TableSelector string
	RowSelector   string
	StartAtHeader bool
}

func RunScrape(ctx context.Context, db *gorm.DB, c *colly.Collector, opts ScrapeOptions) error {

	// ---------------------------
	// Worker pool a BD
	// ---------------------------
	recordsCh := make(chan model.Record, 200)
	errCh := make(chan error, 5)

	// Lanzar workers para insertar a BD
	for i := 0; i < 5; i++ {
		go func(id int) {
			for r := range recordsCh {
				if err := db.Create(&r).Error; err != nil {
					log.Printf("[worker %d] DB insert error: %v", id, err)
					errCh <- err
				}
			}
		}(i)
	}

	// ---------------------------
	// Detectar tabla
	// ---------------------------
	c.OnHTML(opts.TableSelector, func(e *colly.HTMLElement) {
		log.Println("✅ Tabla encontrada:", opts.TableSelector)
	})

	// ---------------------------
	// Procesar filas
	// ---------------------------
	c.OnHTML(opts.TableSelector+" "+opts.RowSelector, func(e *colly.HTMLElement) {

		cells := e.DOM.Find("td")

		if cells.Length() == 0 {
			return
		}

		// Ajustar según la tabla real
		col1 := strings.TrimSpace(cells.Eq(1).Text())  // ClientID
		col2 := strings.TrimSpace(cells.Eq(2).Text())  // Client
		col3 := strings.TrimSpace(cells.Eq(3).Text())  // Date
		col4 := strings.TrimSpace(cells.Eq(6).Text())  // Type
		col5 := strings.TrimSpace(cells.Eq(9).Text())  // Amount
		col6 := strings.TrimSpace(cells.Eq(12).Text()) // Agent

		// Omitir filas sin datos en columna clave
		if col6 == "" {
			return
		}

		rec := model.Record{ // Ajustar según modelo
			ClientID: col1,
			Client:   col2,
			Date:     col3,
			Type:     col4,
			Amount:   col5,
			Agent:    col6,
		}

		log.Printf("✅ Registro: %+v", rec)

		// Enviar a canal de workers la inserción
		select {
		case recordsCh <- rec:
		case <-ctx.Done():
			return
		}
	})

	// ---------------------------
	// Manejo de errores HTTP
	// ---------------------------
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("❌ Error HTTP %s: %v", r.Request.URL, err)
	})

	// ---------------------------
	// Visitar reporte
	// ---------------------------
	log.Println("🌍 Visitando URL:", opts.URL)

	if err := c.Visit(opts.URL); err != nil {
		close(recordsCh)
		return err
	}

	c.Wait()

	// ---------------------------
	// Cerrar workers
	// ---------------------------
	close(recordsCh)
	time.Sleep(1 * time.Second)

	select {
	case e := <-errCh:
		return e
	default:
		return nil
	}
}
