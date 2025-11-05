package scraper

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
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
		colly.UserAgent("go-colly-scraper/1.0"), //
		colly.MaxDepth(2),
	)

	// Ignorar verificación TLS
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	// canal para enviar registros a workers
	recordsCh := make(chan model.Record, 200)
	// canal para errores de inserción
	errCh := make(chan error, 10)

	// worker pool de inserción
	workers := 5
	for i := 0; i < workers; i++ {
		go func(id int) {
			for r := range recordsCh { // recibe registros del canal
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
	c.OnHTML(opts.TableSelector, func(e *colly.HTMLElement) { // e *colly.HTMLElement representa la tabla
		rows := e.DOM.Find(opts.RowSelector)          // e.DOM.Find busca dentro de la tabla
		rows.Each(func(i int, s *goquery.Selection) { // s *goquery.Selection representa cada fila y rows.Each itera sobre ellas
			// aquí puedes procesar cada fila usando s (tipo *goquery.Selection)
		})
	})

	// buscamos cada fila dentro del selector combinado
	c.OnHTML(opts.TableSelector+" "+opts.RowSelector, func(e *colly.HTMLElement) {
		// si la tabla tiene header, saltar la primera fila
		if opts.StartAtHeader && strings.TrimSpace(e.DOM.Parent().Find("tr").First().Text()) == e.Text { //e.DOM.Parent() es la tabla, Find("tr").First().Text() es el texto de la primera fila
			// si coincide con primera fila - heurística: mejor controlar con index en otra forma
		}

		// extraer celdas <td>
		cells := e.DOM.Find("td")
		if cells.Length() == 0 { // cells.Length() == 0 indica que no hay celdas,
			// puede ser header <th>
			return
		}

		// extrae textos de las celdas (ajusta según tabla)
		col1 := strings.TrimSpace(cells.Eq(1).Text())
		col2 := strings.TrimSpace(cells.Eq(2).Text())
		col3 := strings.TrimSpace(cells.Eq(3).Text())
		col4 := strings.TrimSpace(cells.Eq(6).Text())
		col5 := strings.TrimSpace(cells.Eq(9).Text())
		col6 := strings.TrimSpace(cells.Eq(12).Text())

		if col6 == "" {
			return // saltar si columna clave está vacía
		}

		rec := model.Record{ // crea el registro
			ClientID: col1,
			Client:   col2,
			Date:     col3,
			Type:     col4,
			Amount:   col5,
			Agent:    col6,
		}

		// Antes de enviar al canal, imprime los valores
		log.Printf("Fila: ClientID=%s, Client=%s, Date=%s, Type=%s, Amount=%s, Agent=%s",
			col1, col2, col3, col4, col5, col6)

		// enviar registro al canal
		select {
		case recordsCh <- rec: // envía el registro al canal
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
