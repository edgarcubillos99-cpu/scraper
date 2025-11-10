# 📡 Web Scraper Backend – Go + Colly + MySQL + Docker + Goose

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](https://docker.com)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?logo=mysql&logoColor=white)](https://mysql.com)
[![Status](https://img.shields.io/badge/Production_Ready-Yes-brightgreen)]()

Sistema de **scraping backend profesional**, construido con:

✅ **Golang**  
✅ **Colly (sin navegador)**  
✅ **Login automático con CSRF y cookies**  
✅ **Worker Pool concurrente**  
✅ **MySQL con migraciones Goose**  
✅ **Docker y Docker Compose**

El scraper ingresa al portal **UberSmith/Billing**, obtiene el reporte diario, parsea la tabla HTML y guarda los registros en MySQL.

---

# 📑 Índice

- [📌 Descripción General](#-descripción-general)
- [📁 Estructura del Proyecto](#-estructura-del-proyecto)
- [🏗 Arquitectura del Sistema](#-arquitectura-del-sistema)
- [⚙️ Configuración del Entorno](#️-configuración-del-entorno)
- [🐳 Ejecución con Docker](#-ejecución-con-docker)
- [🧠 Diseño del Scraper](#-diseño-del-scraper)
- [🛢 Modelo de Datos](#-modelo-de-datos)
- [📜 Migraciones Goose](#-migraciones-goose)
- [🧪 Pruebas y Verificación](#-pruebas-y-verificación)
- [🚨 Troubleshooting](#-troubleshooting)

---

# 📌 Descripción General

Este proyecto implementa un web scraper backend desarrollado en Golang, utilizando la librería Colly para realizar scraping sin navegador (HTTP puro) y GORM para gestionar la persistencia en una base de datos MySQL.

El sistema está completamente dockerizado, lo que permite su ejecución en cualquier entorno sin necesidad de instalaciones adicionales.

1. Login en **billing.osnetpr.com**
2. Obtención de tokens **CSRF dinámicos**
3. Reescritura correcta de cookies `PHPSESSID`
4. Scraping de tabla HTML estructurada
5. Extracción de campos seleccionados:
   - ClientID  
   - Client  
   - Date  
   - Type  
   - Amount  
   - Agent  
6. Inserciones en MySQL mediante un **worker pool concurrente**
7. Migraciones automáticas antes de ejecutar la app

Todo funciona desde backend, sin navegador, usando únicamente **HTTP/HTML parsing con Colly**.

---

# 📁 Estructura del Proyecto

go-colly-mysql/
├── cmd/
│ └── scraper/main.go
├── internal/
│ ├── db/db.go
│ ├── model/record.go
│ └── scraper/
│ ├── login.go
│ └── scraper.go
├── migrations/
│ ├── 2025..._init_schema.sql
│ ├── 2025..._alter_records_columns.sql
│ └── 2025..._add_indexes.sql
| ├── 2025..._change_date_to_datetime
│ └── 2025..._change_amount_to_decimal
├── docker-compose.yml
├── dockerfile
├── entrypoint.sh
├── goose-db.yaml
├── .env
└── README.md

---

# 🏗 Arquitectura del Sistema

´´´mermaid
flowchart TD

A[Arranque Docker] --> B[MySQL inicia y pasa healthcheck]
B --> C[entrypoint.sh]
C --> D[Goose ejecuta migraciones]
D --> E[Lanzar binario scraper]

E --> F[InitCollector()]
F --> G[Login() -> CSRF + Cookies]
G --> H[Construcción URL con timestamps]

H --> I[RunScrape()]
I --> J[Parseo de tabla HTML]
J --> K[Worker Pool]
K --> L[Inserción MySQL]
L --> M[Finaliza ejecución]

---

# ⚙️ Configuración del Entorno

1) .env
 Credenciales de la pagina de login
BILLING_USER=usuario_demo
BILLING_PASS=pass_demo

 Horario de ejecución del scraper
TIMEZONE=America/Puerto_Rico
SCRAPER_HOUR=HORA_EJECUCION_0-23
SCRAPER_MINUTE=MINUTO_EJECUCION_0-59

2) docker-compose
DB_HOST=mysql
DB_PORT=3306
DB_USER=app
DB_PASSWORD=apppass
DB_NAME=appdb

TARGET_URL=https://billing.osnetpr.com/admin/reports/report_billings_detail.php?type=&begin=%d&end=%d&type=
TABLE_SELECTOR=table.table-body-modern
ROW_SELECTOR=tr

---

# 🐳 Ejecución con Docker

Ejecutar: docker compose up --build

El proceso:

✅ 1. MySQL inicia con healthcheck
✅ 2. Goose aplica migraciones
✅ 3. Se lanza /root/scraper
✅ 4. El login automático se ejecuta
✅ 5. Se obtiene el reporte del día
✅ 6. Se parsean filas y se insertan en la BD

Logs: docker logs -f app

---

# 🧠 Diseño del Scraper
✅ Login (login.go)

Se hace GET inicial al portal

Se extraen csrf_key y csrf_token desde los <script>

Se envía POST completo

Se verifica cookie PHPSESSID

Se reescribe cookie → dominio raíz (billing.osnetpr.com)

Esto permite mantener la sesión en cualquier subruta del portal.

🛠 RunScrape (scraper.go)

Detecta tabla vía TABLE_SELECTOR

Recorre tr de la tabla

Extrae columnas:
td[1] → ClientID
td[2] → Client
td[3] → Date
td[6] → Type
td[9] → Amount
td[12] → Agent

Envía cada registro a un canal bufferizado (200)

Worker pool de 5 goroutines inserta en MySQL

---

# 🛢 Modelo de Datos
Archivo: internal/model/record.go
type Record struct {
    gorm.Model
    ClientID string
    Client   string
    Date     string
    Type     string
    Amount   string
    Agent    string
}

---

# 📜 Migraciones Goose

Migraciones en /migrations:

✅ init_schema.sql
Crea tabla records.

✅ alter_records_columns.sql
Ajusta tamaños de columnas.

✅ add_indexes.sql
Añade índices para acelerar consultas.

✅ change_date_to_datetime
Cambia el tipo de dato a date

✅ change_amount_to_decimal
Cambia el tipo de dato decimal

Se ejecutan automáticamente por: entrypoint.sh → goose up

---

# 🧪 Pruebas y Verificación
✅ Entrar al contenedor MySQL
docker exec -it scraper-mysql-1 mysql -uapp -papppass appdb

✅ Ver datos insertados
SELECT * FROM records;

✅ Ver errores desde el worker pool
docker logs -f app

---

# 🚨 Troubleshooting
❌ Error: no se encuentra PHPSESSID después del POST

Revisar credenciales en .env

Portal pudo haber cambiado CSRF

Ver logs de csrf_key y csrf_token

❌ Tabla no encontrada

Goose no ejecutó migraciones → verificar: docker compose logs app

❌ Scraping devuelve filas vacías

Revisar TABLE_SELECTOR

Revisar índices td[x]

Ver página HTML real usando: docker exec -it app sh
curl <URL>


