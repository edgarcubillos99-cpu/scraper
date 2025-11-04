package model

import "gorm.io/gorm"

// Record representa una fila de la tabla que scrapeamos.
// Ajusta campos de columnas según la tabla real.
type Record struct {
	gorm.Model
	Col1 string `gorm:"type:varchar(255);index"`
	Col2 string `gorm:"type:varchar(255)"`
	Col3 string `gorm:"type:text"`
}
