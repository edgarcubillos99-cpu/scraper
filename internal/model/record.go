package model

import "gorm.io/gorm"

// Record representa una fila de la tabla que scrapeamos.
// Ajusta campos de columnas según la tabla real.
type Record struct {
	gorm.Model
	ClientID string `gorm:"column:client_id"`
	Client   string `gorm:"column:client"`
	Date     string `gorm:"column:date"`
	Type     string `gorm:"column:type"`
	Amount   string `gorm:"column:amount"`
	Agent    string `gorm:"column:agent"`
}
