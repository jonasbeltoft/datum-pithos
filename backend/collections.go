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
Deletes a collection from the collections table

Query params:

	collection_id: int
*/
func deleteCollectionHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	_collectionId := r.FormValue("collection_id")

	// Validate input
	collectionId, err := strconv.Atoi(_collectionId)
	if err != nil {
		http.Error(w, "collection_id must be a positive int", http.StatusBadRequest)
		return
	}

	// Delete collection from database
	query := "DELETE FROM collections WHERE id = ?"
	result, err := DB.Exec(query, collectionId)
	if err != nil {
		http.Error(w, "failed to delete collection", http.StatusInternalServerError)
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "collection not found", http.StatusNotFound)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
}

/*
Gets one or more collections from db

Query params:

	id: int

Result:

	[{
		id: int,
		name: string,
		description: string
	}]
*/
func fetchCollectionsHandler(w http.ResponseWriter, r *http.Request) {
	var collections = []Collection{}
	// Get the id of the collection
	_id := r.FormValue("id")
	if _id == "" {
		// No id, so get all
		rows, err := DB.Query("SELECT id, name, description FROM collections")
		if err != nil {
			http.Error(w, "error when reading from database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var collection Collection
			if err := rows.Scan(&collection.Id, &collection.Name, &collection.Description); err != nil {
				http.Error(w, "error when reading from database", http.StatusInternalServerError)
				return
			}
			collections = append(collections, collection)
		}

		// Check for errors from iterating over rows
		if err = rows.Err(); err != nil {
			http.Error(w, "error when reading from database", http.StatusInternalServerError)
			return
		}

	} else {
		// Fetch specific collection
		id, err := strconv.Atoi(_id)
		if err != nil || id < 1 {
			http.Error(w, "id must be a positive int", http.StatusBadRequest)
			return
		}
		// Get from DB
		collection := Collection{
			Id: &id,
		}
		err = DB.QueryRow("SELECT name, description FROM collections WHERE id = ?", id).Scan(&collection.Name, &collection.Description)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNoContent)
				fmt.Fprintln(w)
				return
			}
			http.Error(w, "error when reading from database", http.StatusInternalServerError)
			return
		}
		collections = append(collections, collection)
	}
	result, err := json.Marshal(collections)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(result))
}

/*
Inserts a new collection into the collections table

Body:

	{
		name: string,
		description: string
	}
*/
func insertCollectionHandler(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request
	var collection Collection
	if err := json.NewDecoder(r.Body).Decode(&collection); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate required fields
	if collection.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if collection.Description == nil {
		collection.Description = new(string)
		*collection.Description = ""
	}

	// Insert into database
	query := "INSERT INTO collections (name, description) VALUES (?, ?)"
	result, err := DB.Exec(query, collection.Name, collection.Description)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: collections.name") {
			http.Error(w, "collection already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to insert collection", http.StatusInternalServerError)
		return
	}

	// Get the inserted ID
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

// Collection represents the structure of the collection table
type Collection struct {
	Id          *int    `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}
