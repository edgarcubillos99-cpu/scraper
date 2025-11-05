package db

import (
	"fmt"
	"log"
	"os"

	"github.com/eddgaroso/go-colly-mysql/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connect establece la conexión a la base de datos MySQL usando GORM.
func Connect() (*gorm.DB, error) {
	// os.Getenv lee las variables de entorno para la configuración de la base de datos
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	// construir DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{}) // abrir conexión
	if err != nil {
		return nil, err
	}

	// Auto-migrate (crea/actualiza tabla)
	// Desactivar en producción para evitar cambios no deseados en el esquema
	if err := db.AutoMigrate(&model.Record{}); err != nil { // migrar modelo Record
		log.Println("AutoMigrate error:", err)
		return nil, err
	}

	return db, nil
}
