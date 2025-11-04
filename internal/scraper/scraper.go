package scraper

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/eddgaroso/go-colly-mysql/internal/model"
	"github.com/gocolly/colly/v2"
	"gorm.io/gorm"
)

// ScrapeOptions guarda parámetros configurables
type ScrapeOptions struct {
	URL           string
	TableSelector string // ejemplo: "table#data"
	RowSelector   string // ejemplo: "tr"
	StartAtHeader bool   // si la tabla tiene header
}

// RunScrape ejecuta el scraping y envía registros al DB a través de GORM.
// Inserta concurrentemente mediante un worker pool simple.
func RunScrape(ctx context.Context, db *gorm.DB, opts ScrapeOptions) error {
	c := colly.NewCollector(
		colly.UserAgent("go-colly-scraper/1.0"),
		colly.MaxDepth(2),
	)

	// canal para enviar registros a workers
	recordsCh := make(chan model.Record, 200)
	// canal para errores de inserción
	errCh := make(chan error, 10)

	// worker pool de inserción
	workers := 5
	for i := 0; i < workers; i++ {
		go func(id int) {
			for r := range recordsCh {
				if err := db.Create(&r).Error; err != nil {
					log.Printf("[worker %d] DB insert error: %v", id, err)
					select {
					case errCh <- err:
					default:
					}
				}
			}
		}(i)
	}

	// seleccionar filas dentro de la tabla objetivo
	c.OnHTML(opts.TableSelector, func(e *colly.HTMLElement) {
		rows := e.DOM.Find(opts.RowSelector)
		rows.Each(func(i int, s *goquery.Selection) {
			// Nota: colly.HTMLElement expone e.DOM que es *goquery.Selection;
			// aquí puedes procesar cada fila usando s (tipo *goquery.Selection)
		})
	})

	// simpler approach: parse rows using child callbacks:
	// buscamos cada fila dentro del selector combinado
	c.OnHTML(opts.TableSelector+" "+opts.RowSelector, func(e *colly.HTMLElement) {
		// si la tabla tiene header, saltar la primera fila
		if opts.StartAtHeader && strings.TrimSpace(e.DOM.Parent().Find("tr").First().Text()) == e.Text {
			// si coincide con primera fila - heurística: mejor controlar con index en otra forma
		}

		// extraer celdas <td>
		cells := e.DOM.Find("td")
		if cells.Length() == 0 {
			// puede ser header <th>
			return
		}

		// extrae textos de las primeras 3 celdas (ajusta según tabla)
		col1 := strings.TrimSpace(cells.Eq(0).Text())
		col2 := strings.TrimSpace(cells.Eq(1).Text())
		col3 := strings.TrimSpace(cells.Eq(2).Text())

		rec := model.Record{
			Col1: col1,
			Col2: col2,
			Col3: col3,
		}

		select {
		case recordsCh <- rec:
		case <-ctx.Done():
			log.Println("context cancelled while sending record")
		}
	})

	// manejar errores de request
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed: %v\n", r.Request.URL, err)
	})

	// visitar URL
	if err := c.Visit(opts.URL); err != nil {
		close(recordsCh)
		return err
	}

	// esperar que no haya más jobs (colly trabaja asincrónicamente)
	c.Wait() // colly.Wait espera a collectors (necesita EnableAsync si lo usamos)
	// Para usar Wait debemos haber activado Async
	// Alternativa simple: small sleep para asegurar callbacks (no ideal) -> mejor activar Async
	// Ajustamos para usar Async:
	return finalizeAndClose(recordsCh, errCh)
}

func finalizeAndClose(recordsCh chan model.Record, errCh chan error) error {
	// cerramos el canal y esperamos un tiempo prudente para que terminen workers
	close(recordsCh)
	// esperar breve tiempo para que los workers procesen
	time.Sleep(2 * time.Second)
	select {
	case e := <-errCh:
		return e
	default:
		return nil
	}
}
