package model

import "gorm.io/gorm"

// Record representa una fila de la tabla que scrapeamos.
// Ajusta campos de columnas según la tabla real.
type Record struct {
	gorm.Model
	ClientID string `gorm:"type:varchar(50);column:client_id"`
	Client   string `gorm:"type:varchar(100);column:client"`
	Date     string `gorm:"type:varchar(20);column:date"`
	Type     string `gorm:"type:varchar(50);column:type"`
	Amount   string `gorm:"type:varchar(50);column:amount"`
	Agent    string `gorm:"type:varchar(100);column:agent"`
}
