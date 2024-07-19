package main

import (
	"encoding/csv"
	"fmt"
	"ikimeia_stream/m/set"
	"os"
)


func UploadCSV(filePath string ) error {

	// Exécuter une requête SQL pour extraire les données
	rows, err := db.Query("SELECT * FROM tokens")
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	// Obtenir les colonnes de la requête
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("Failed to get columns: %v", err)
	}
    
    file, err := os.Create(filePath)
    if err!= nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
	defer writer.Flush()
    // Écrire les en-têtes des colonnes dans le fichier CSV
	if err := writer.Write(columns); err != nil {
		log.Fatalf("Failed to write headers to CSV: %v", err)
	}

	// Créer une slice d'interfaces pour stocker les valeurs de chaque ligne
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Lire les lignes et les écrire dans le fichier CSV
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		record := make([]string, len(columns))
		for i, value := range values {
			if value != nil {
				record[i] = fmt.Sprintf("%v", value)
			}
		}

		if err := writer.Write(record); err != nil {
			log.Fatalf("Failed to write record to CSV: %v", err)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

    set.PrintSomething("success")
    return nil
}
