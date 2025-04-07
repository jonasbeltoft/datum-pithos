package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

/*
Inserts a new unit into the units table

Query params:

	name: string
*/
func insertUnitHandler(w http.ResponseWriter, r *http.Request) {
	// Get the name of the unit
	name := r.FormValue("name")

	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	} else if len(name) > 32 {
		http.Error(w, "name must be at most 32 characters", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO units (name) VALUES (?)"
	result, err := DB.Exec(query, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: units.name") {
			http.Error(w, "unit already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "error when inserting unit to database", http.StatusInternalServerError)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		// return only StatusOk
		fmt.Fprintln(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{\"id\":  %d}", id)
}

/*
Gets one or more units from db

Query params:

	id: int

Result:

	[{
		id: int,
		name: string
	}]
*/
func fetchUnitsHandler(w http.ResponseWriter, r *http.Request) {
	var units = []Unit{}
	// Get the id of the unit
	_id := r.FormValue("id")
	if _id == "" {
		// No id, so get all
		rows, err := DB.Query("SELECT id, name FROM units")
		if err != nil {
			http.Error(w, "error when reading from database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var unit Unit
			if err := rows.Scan(&unit.Id, &unit.Name); err != nil {
				http.Error(w, "error when reading from database", http.StatusInternalServerError)
				return
			}
			units = append(units, unit)
		}

		// Check for errors from iterating over rows
		if err = rows.Err(); err != nil {
			http.Error(w, "error when reading from database", http.StatusInternalServerError)
			return
		}

	} else {
		// Fetch specific unit
		id, err := strconv.Atoi(_id)
		if err != nil || id < 1 {
			http.Error(w, "id must be a positive int", http.StatusBadRequest)
			return
		}
		// Get from DB
		unit := Unit{
			Id: id,
		}
		err = DB.QueryRow("SELECT name FROM units WHERE id = ?", id).Scan(&unit.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNoContent)
				fmt.Fprintln(w)
				return
			}
			http.Error(w, "error when reading from database", http.StatusInternalServerError)
			return
		}
		units = append(units, unit)
	}
	result, err := json.Marshal(units)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(result))
}

type Unit struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
