package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

/*
Deletes an attribute from a collection

Query params:

	attribute_id: int
*/
func deleteAttributesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	_attributeId := r.FormValue("attribute_id")

	// Validate input
	attributeId, err := strconv.Atoi(_attributeId)
	if err != nil {
		http.Error(w, "attribute_id must be a positive int", http.StatusBadRequest)
		return
	}

	// Delete attribute from database
	query := "DELETE FROM sample_attributes WHERE id = ?"
	result, err := DB.Exec(query, attributeId)
	if err != nil {
		http.Error(w, "failed to delete attribute", http.StatusInternalServerError)
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "attribute not found", http.StatusNotFound)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
}

/*
Inserts a new attribute for a collection

Query params:

	collection_id: int,
	name: string,
	unit_id?: int
*/
func insertAttributesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req = insertAttributeRequest{}
	_collectionId := r.FormValue("collection_id")
	req.Name = r.FormValue("name")
	_unitId := r.FormValue("unit_id")

	// Validate input
	var err error
	req.CollectionId, err = strconv.Atoi(_collectionId)
	if err != nil || req.Name == "" {
		http.Error(w, "collection_id and name are required", http.StatusBadRequest)
		return
	}
	if _unitId != "" {
		_id, err := strconv.Atoi(_unitId)
		if err != nil {
			http.Error(w, "invalid unit_id", http.StatusBadRequest)
			return
		}
		req.UnitId = &_id
	}

	// Insert attribute into database
	var query string
	var args []interface{}

	if req.UnitId != nil {
		query = "INSERT INTO sample_attributes (collection_id, name, unit_id) VALUES (?, ?, ?)"
		args = append(args, req.CollectionId, req.Name, *req.UnitId)
	} else {
		query = "INSERT INTO sample_attributes (collection_id, name) VALUES (?, ?)"
		args = append(args, req.CollectionId, req.Name)
	}

	result, err := DB.Exec(query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			http.Error(w, "atttribute already exists on this collection", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to insert attribute", http.StatusInternalServerError)
		return
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		// return only StatusOk
		fmt.Fprintln(w)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"attribute_id": int(id)})
}

type insertAttributeRequest struct {
	CollectionId int    `json:"collection_id"`
	Name         string `json:"name"`
	UnitId       *int   `json:"unit_id,omitempty"` // Nullable
}
