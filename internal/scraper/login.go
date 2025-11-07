package scraper

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Datos de login
type LoginInfo struct {
	Username string
	Password string
}

// Inicializa collector con CookieJar y TLS ignorado
func InitCollector() (*colly.Collector, *cookiejar.Jar, error) {
	jar, _ := cookiejar.New(nil)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	c := colly.NewCollector(
		colly.MaxDepth(10),
		colly.UserAgent("Mozilla/5.0"),
		colly.AllowURLRevisit(),
		colly.AllowedDomains("billing.osnetpr.com"),
	)

	// permitir redirects (UberSmith usa muchos)
	c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		return nil
	})

	// habilitar uso de TLS inseguro y cookies
	c.WithTransport(transport)
	c.SetCookieJar(jar)

	return c, jar, nil
}

// LOGIN con extracción de CSRF automática
func Login(c *colly.Collector, jar *cookiejar.Jar, creds LoginInfo) error {

	loginURL := "https://billing.osnetpr.com/admin/"

	// Variables donde guardaremos los tokens CSRF
	var csrfKey string
	var csrfToken string

	// Capturar CSRF desde scripts del login
	c.OnHTML("script", func(e *colly.HTMLElement) {
		text := e.Text

		// comprobar si el script contiene las variables CSRF
		if strings.Contains(text, "csrf_key") {
			reKey := regexp.MustCompile(`csrf_key\s*=\s*'([^']+)'`)
			reTok := regexp.MustCompile(`csrf_token\s*=\s*'([^']+)'`)

			if match := reKey.FindStringSubmatch(text); len(match) == 2 {
				csrfKey = match[1]
			}
			if match := reTok.FindStringSubmatch(text); len(match) == 2 {
				csrfToken = match[1]
			}
		}
	})

	// Hacer GET inicial para obtener cookies y CSRF
	log.Println("✅ GET inicial para obtener cookies y CSRF...")
	if err := c.Visit(loginURL); err != nil {
		return fmt.Errorf("error GET inicial: %w", err)
	}
	c.Wait()

	log.Printf("✅ CSRF capturado: key=%s token=%s\n", csrfKey, csrfToken)

	if csrfKey == "" || csrfToken == "" {
		return fmt.Errorf("❌ no se pudo extraer CSRF del login")
	}

	// Preparar el formulario con CSRF correcto
	form := map[string]string{
		"login":           creds.Username,
		"pass":            creds.Password,
		"logclick":        "Login",
		"uber_csrf_key":   csrfKey,
		"uber_csrf_token": csrfToken,
	}

	log.Println("Enviando POST de login...")

	//Hacer POST de login
	if err := c.Post(loginURL, form); err != nil {
		return fmt.Errorf("error POST login: %w", err)
	}
	c.Wait()

	// Validar cookie PHPSESSID
	u, _ := url.Parse(loginURL)
	cookies := jar.Cookies(u)

	// Buscar PHPSESSID
	var session string
	for _, ck := range cookies {
		if ck.Name == "PHPSESSID" {
			session = ck.Value
			log.Printf("Cookie: %s=%s", ck.Name, ck.Value)
		}
	}

	if session == "" {
		return fmt.Errorf("❌ login falló: no se encontró PHPSESSID después del POST")
	}

	log.Println("✅ Login exitoso — PHPSESSID presente.")

	//Reparar cookie para que funcione en todas las rutas
	root, _ := url.Parse("https://billing.osnetpr.com/")
	php := &http.Cookie{
		Name:   "PHPSESSID",
		Value:  session,
		Domain: "billing.osnetpr.com",
		Path:   "/",
	}

	jar.SetCookies(root, []*http.Cookie{php})

	log.Printf("✅ COOKIE REPARADA: PHPSESSID=%s (domain=%s, path=%s)",
		session, php.Domain, php.Path)

	return nil
}
