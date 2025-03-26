package main

import (
	"database/sql"
	"fmt"
)

// InsertUnit inserts a new unit into the units table
func InsertUnit(db *sql.DB, name string) (int64, error) {
	query := "INSERT INTO units (name) VALUES (?)"
	result, err := db.Exec(query, name)
	if err != nil {
		return 0, fmt.Errorf("InsertUnit: %w", err)
	}
	return result.LastInsertId()
}
