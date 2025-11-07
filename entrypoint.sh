#!/bin/sh
set -e

echo "✅ Ejecutando migraciones de Goose..."

# Construir cadena DSN de MySQL desde variables de entorno del container
DB_DSN="${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true"

# Ejecutar migraciones
goose -dir /migrations mysql "$DB_DSN" up

echo "✅ Migraciones aplicadas. Ejecutando scraper..."
/root/scraper
