package scraper

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/gocolly/colly/v2"
)

// Credenciales del usuario
type LoginInfo struct {
	Username string
	Password string
}

// Crea un collector con CookieJar y TLS ignorado (certificado self-signed)
func InitCollector() (*colly.Collector, *cookiejar.Jar, error) {
	jar, _ := cookiejar.New(nil)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.UserAgent("Mozilla/5.0"),
		colly.MaxDepth(3),
	)

	c.WithTransport(transport)
	c.SetCookieJar(jar)

	return c, jar, nil
}

// Login real y funcional
func Login(c *colly.Collector, jar *cookiejar.Jar, creds LoginInfo) error {

	loginURL := "https://billing.osnetpr.com/admin/"

	// 1) GET inicial para obtener PHPSESSID
	log.Println("✅ GET inicial para obtener cookies...")
	if err := c.Visit(loginURL); err != nil {
		return fmt.Errorf("error GET inicial: %w", err)
	}
	c.Wait()

	// 2) Hacer POST (Colly enviará cookies automáticamente)
	form := map[string]string{
		"login":           creds.Username,
		"pass":            creds.Password,
		"logclick":        "Login",
		"uber_csrf_key":   "",
		"uber_csrf_token": "",
	}

	log.Println("✅ Enviando POST de login...")
	if err := c.Post(loginURL, form); err != nil {
		return fmt.Errorf("error POST login: %w", err)
	}
	c.Wait()

	// 3) Validar que el login funcionó revisando PHPSESSID
	u, _ := url.Parse(loginURL)
	cookies := jar.Cookies(u)

	hasSession := false
	for _, ck := range cookies {
		log.Printf("Cookie: %s=%s", ck.Name, ck.Value)
		if ck.Name == "PHPSESSID" && ck.Value != "" {
			hasSession = true
		}
	}

	if !hasSession {
		return fmt.Errorf("❌ login falló: no se encontró PHPSESSID después del POST")
	}

	log.Println("✅ Login exitoso — PHPSESSID presente.")
	return nil
}
