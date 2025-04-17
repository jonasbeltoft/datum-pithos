package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
Updates a single value in a sample

Query params:

	sample_id: int,
	attribute_id: int,
	value: string
*/
func insertOrUpdateSampleValueHandler(w http.ResponseWriter, r *http.Request) {
	// Get data from query parameters
	_attributeId := r.FormValue("attribute_id")
	_sampleId := r.FormValue("sample_id")
	value := r.FormValue("value")
	if _attributeId == "" || _sampleId == "" {
		http.Error(w, "attribute_id and sample_id is required", http.StatusBadRequest)
		return
	}

	if len(value) > 32 {
		http.Error(w, "value must be at most 32 characters", http.StatusBadRequest)
		return
	}

	// Convert sample_id and attribute_id to int
	sampleId, err := strconv.Atoi(_sampleId)
	if err != nil {
		http.Error(w, "sample_id must be a positive int", http.StatusBadRequest)
		return
	}
	attributeId, err := strconv.Atoi(_attributeId)
	if err != nil {
		http.Error(w, "attribute_id must be a positive int", http.StatusBadRequest)
		return
	}

	queryInsert := "INSERT INTO sample_attribute_values (sample_id, attribute_id, value) VALUES (?, ?, ?)"
	result, err := DB.Exec(queryInsert, sampleId, attributeId, value)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			// If the sample already exists, update it instead
			query := "UPDATE sample_attribute_values SET value = ? WHERE sample_id = ? AND attribute_id = ?"
			update_result, err := DB.Exec(query, value, sampleId, attributeId)
			if err != nil {
				http.Error(w, "error when updating sample value in database", http.StatusInternalServerError)
				return
			}
			rowsAffected, err := update_result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				http.Error(w, "sample not found or no changes made", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "{\"status\": \"success\"}")
			return

		} else {
			http.Error(w, "error when inserting sample value to database", http.StatusInternalServerError)
			return
		}
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "sample not found or no changes made", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "{\"status\": \"success\"}")
}

/*
Updates a sample in the db

Query params:

	sample_id: int,
	created_at: int, // UNIX timestamp in seconds
	note: string,
*/
func updateSampleHandler(w http.ResponseWriter, r *http.Request) {
	// Get data from query parameters
	_sampleId := r.FormValue("sample_id")
	_createdAt := r.FormValue("created_at")
	note := r.FormValue("note")
	if _sampleId == "" {
		http.Error(w, "sample_id is required", http.StatusBadRequest)
		return
	}
	if !strings.Contains(r.URL.String(), "note") && _createdAt == "" {
		http.Error(w, "note or created_at is required", http.StatusBadRequest)
		return
	}

	query := "UPDATE samples SET"
	args := []interface{}{}

	if strings.Contains(r.URL.String(), "note") {
		query += " note = ?,"
		args = append(args, note)
	}
	// Convert sample_id and created_at to int
	var createdAt int64
	if _createdAt != "" {
		time, err := strconv.Atoi(_createdAt)
		if err != nil {
			http.Error(w, "created_at must be a positive int", http.StatusBadRequest)
			return
		}
		createdAt = int64(time)
		query += " created_at = ?,"
		args = append(args, createdAt)
	}
	sampleId, err := strconv.Atoi(_sampleId)
	if err != nil {
		http.Error(w, "sample_id must be a positive int", http.StatusBadRequest)
		return
	}
	args = append(args, sampleId)

	// Update sample in the database
	query = strings.TrimSuffix(query, ",") + " WHERE id = ?"
	result, err := DB.Exec(query, args...)
	if err != nil {
		http.Error(w, "error when updating sample in database", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "value not found or no changes made", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "{\"status\": \"success\"}")
}

/*
Deletes a sample from the db

Query params:

	sample_id: int
*/
func deleteSampleHandler(w http.ResponseWriter, r *http.Request) {
	// Get data from query parameters
	_sampleId := r.FormValue("sample_id")
	if _sampleId == "" {
		http.Error(w, "sample_id is required", http.StatusBadRequest)
		return
	}

	// Convert sample_id to int
	sampleId, err := strconv.Atoi(_sampleId)
	if err != nil {
		http.Error(w, "sample_id must be a positive int", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM samples WHERE id = ?"
	result, err := DB.Exec(query, sampleId)
	if err != nil {
		http.Error(w, "error when deleting sample from database", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "sample not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "{\"status\": \"success\"}")
}

/*
Gets one or more samples from db

Query params:

	collection_id: int,
	sample_id: int,
	page_size: int,
	page: int,
	before: int, // UNIX timestamp in seconds
	after: int // UNIX timestamp in seconds

Result:

	{
		attributes: [
			{
				attribute_id: int,
				name: string
				unit_id: int
			}
			...
		],
		samples: [
			{
				sample_id: int,
				note: string,
				created_at: int,
				values: [
					{
						attribute_id: int,
						value: string
					}
					...
				]
			}
			...
		],
		total_count: int,
		page_info: {
			limit: int,
			offset: int,
			has_next_page: true
		}
	}
*/
func fetchSamplesHandler(w http.ResponseWriter, r *http.Request) {
	// Get data from query parameters
	_collectionId := r.FormValue("collection_id")
	_sampleId := r.FormValue("sample_id")
	if _collectionId == "" && _sampleId == "" {
		http.Error(w, "collection_id or sample_id is required", http.StatusBadRequest)
		return
	}
	var sampleId int
	singleSample := false
	if _sampleId != "" {
		id, err := strconv.Atoi(_sampleId)
		if err != nil {
			http.Error(w, "sample_id must be a positive int", http.StatusBadRequest)
			return
		}
		singleSample = true
		sampleId = id
	}
	var collectionId int
	if _collectionId != "" {
		id, err := strconv.Atoi(_collectionId)
		if err != nil {
			http.Error(w, "collection_id must be a positive int", http.StatusBadRequest)
			return
		}
		collectionId = id
	}

	// Check if collectionId is valid
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM collections WHERE id = ?)`
	err := DB.QueryRow(query, collectionId).Scan(&exists)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}

	// Parse pagination params
	_pageSize := r.FormValue("page_size")
	_page := r.FormValue("page")
	pageSize := 0 // Default limit
	page := 0     // Default offset
	var sampleArgs []interface{}
	sampleArgs = append(sampleArgs, collectionId)

	if _pageSize != "" {
		l, err := strconv.Atoi(_pageSize)
		if err != nil || l <= 0 {
			http.Error(w, "page_size must be a positive integer", http.StatusBadRequest)
			return
		}
		pageSize = l
	}
	if _page != "" {
		o, err := strconv.Atoi(_page)
		if err != nil || o <= 0 {
			http.Error(w, "page must be a positive integer", http.StatusBadRequest)
			return
		}
		page = o
	}
	// Calculate paging offset
	offset := pageSize * (page - 1)

	// Parse time filter
	_before := r.FormValue("before")
	_after := r.FormValue("after")
	before := 0 // Default limit
	after := 0  // Default offset

	if _before != "" {
		l, err := strconv.Atoi(_before)
		if err != nil || l <= 0 {
			http.Error(w, "before must be a positive integer", http.StatusBadRequest)
			return
		}
		before = l
	}
	if _after != "" {
		o, err := strconv.Atoi(_after)
		if err != nil || o <= 0 {
			http.Error(w, "after must be a positive integer", http.StatusBadRequest)
			return
		}
		after = o
	}

	// Fetch attributes related to this collection
	attributeQuery := "SELECT id, name, unit_id FROM sample_attributes WHERE collection_id = ? ORDER BY id"
	attrRows, err := DB.Query(attributeQuery, collectionId)
	if err != nil {
		log.Println("DB Fetch Error:", err)
		http.Error(w, "Failed to fetch attributes", http.StatusInternalServerError)
		return
	}
	defer attrRows.Close()

	var attributes = make([]Attribute, 0)
	for attrRows.Next() {
		var attr = Attribute{}
		if err := attrRows.Scan(&attr.AttributeId, &attr.Name, &attr.UnitId); err != nil {
			http.Error(w, "Error reading attributes", http.StatusInternalServerError)
			return
		}
		attributes = append(attributes, attr)
	}
	// Check for row iteration errors
	if err = attrRows.Err(); err != nil {
		http.Error(w, "Error iterating attribute rows", http.StatusInternalServerError)
		return
	}

	var samples = make([]Sample, 0)
	var totalCount int

	if singleSample {
		var sample = Sample{
			SampleId: sampleId,
		}
		// Fetch sample
		sampleQuery := "SELECT created_at, note FROM samples WHERE id = ?"
		err := DB.QueryRow(sampleQuery, sampleId).Scan(&sample.CreatedAt, &sample.Note)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNoContent)
				fmt.Fprintln(w)
				return
			}
			http.Error(w, "error when reading from database", http.StatusInternalServerError)
			return
		}

		// Fetch values associated with this sample
		valueQuery := "SELECT attribute_id, value FROM sample_attribute_values WHERE sample_id = ?"
		valueRows, err := DB.Query(valueQuery, sample.SampleId)
		if err != nil {
			http.Error(w, "Failed to fetch sample values", http.StatusInternalServerError)
			return
		}
		defer valueRows.Close()

		var values = make([]SampleValue, 0)
		for valueRows.Next() {
			var val SampleValue
			if err := valueRows.Scan(&val.AttributeId, &val.Value); err != nil {
				http.Error(w, "Error reading values", http.StatusInternalServerError)
				return
			}
			values = append(values, val)
		}

		sample.Values = values
		samples = append(samples, sample)

		// Check for row iteration errors
		if err = valueRows.Err(); err != nil {
			http.Error(w, "Error iterating sample rows", http.StatusInternalServerError)
			return
		}
	} else {
		// Fetch samples related to this collection

		sampleQuery := "SELECT id, created_at, note FROM samples WHERE collection_id = ?"

		// If filtering, add args
		if before != 0 {
			sampleQuery += " AND created_at < ?"
			sampleArgs = append(sampleArgs, before)
		}
		if _after != "" {
			sampleQuery += " AND created_at > ?"
			sampleArgs = append(sampleArgs, after)
		}

		// If paging, add args
		sampleQuery += " ORDER BY created_at DESC"
		if pageSize > 0 && page > 0 {
			sampleQuery += " LIMIT ? OFFSET ?"
			sampleArgs = append(sampleArgs, pageSize, offset)
		}
		// Do the query
		sampleRows, err := DB.Query(sampleQuery, sampleArgs...)
		if err != nil {
			http.Error(w, "Failed to fetch samples", http.StatusInternalServerError)
			log.Println("DB Fetch Error (Samples):", err)
			return
		}
		defer sampleRows.Close()

		for sampleRows.Next() {
			var sample Sample
			if err := sampleRows.Scan(&sample.SampleId, &sample.CreatedAt, &sample.Note); err != nil {
				http.Error(w, "Error reading samples", http.StatusInternalServerError)
				return
			}

			// Fetch values associated with each sample
			valueQuery := "SELECT attribute_id, value FROM sample_attribute_values WHERE sample_id = ?"
			valueRows, err := DB.Query(valueQuery, sample.SampleId)
			if err != nil {
				http.Error(w, "Failed to fetch sample values", http.StatusInternalServerError)
				log.Println("DB Fetch Error (Values):", err)
				return
			}
			defer valueRows.Close()

			var values = make([]SampleValue, 0)
			for valueRows.Next() {
				var val SampleValue
				if err := valueRows.Scan(&val.AttributeId, &val.Value); err != nil {
					http.Error(w, "Error reading values", http.StatusInternalServerError)
					return
				}
				values = append(values, val)
			}

			sample.Values = values
			samples = append(samples, sample)
		}

		// Check for row iteration errors
		if err = sampleRows.Err(); err != nil {
			http.Error(w, "Error iterating sample rows", http.StatusInternalServerError)
			return
		}

		// Count total samples for pagination metadata
		countQuery := strings.Replace(sampleQuery, "SELECT id, created_at, note FROM", "SELECT COUNT(*) FROM", 1)
		if strings.Contains(countQuery, " LIMIT ? OFFSET ?") {
			countQuery = strings.TrimSuffix(countQuery, " LIMIT ? OFFSET ?")
			sampleArgs = sampleArgs[:len(sampleArgs)-2]
		}

		err = DB.QueryRow(countQuery, sampleArgs...).Scan(&totalCount)
		if err != nil {
			http.Error(w, "Failed to count total samples", http.StatusInternalServerError)
			return
		}
	}

	// Return JSON response
	response := FetchSamplesResponse{
		Attributes: attributes,
		Samples:    samples,
		TotalCount: totalCount,
		PageInfo: PageInfo{
			Limit:       pageSize,
			Offset:      offset,
			HasNextPage: page != 0 && offset+pageSize < totalCount,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Attribute represents an attribute of a sample
type Attribute struct {
	AttributeId int    `json:"attribute_id"`
	Name        string `json:"name"`
	UnitId      *int   `json:"unit_id,omitempty"`
}

// Sample represents a sample entry in the response
type Sample struct {
	SampleId  int           `json:"sample_id"`
	Note      string        `json:"note"`
	CreatedAt int64         `json:"created_at"`
	Values    []SampleValue `json:"values"`
}

// FetchSamplesResponse represents the full response
type FetchSamplesResponse struct {
	Attributes []Attribute `json:"attributes"`
	Samples    []Sample    `json:"samples"`
	TotalCount int         `json:"total_count,omitempty"`
	PageInfo   PageInfo    `json:"page_info"`
}

// Pagination metadata
type PageInfo struct {
	Limit       int  `json:"limit"`
	Offset      int  `json:"offset"`
	HasNextPage bool `json:"has_next_page"`
}

/*
Inserts a sample into the db

Body:

	{
		collection_id: int,
		note: string,
		values: [
			{
				attribute_id: int,
				value: string
			}
			...
		]
	}
*/
func insertSampleHandler(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request
	var sample InsertSampleBody
	if err := json.NewDecoder(r.Body).Decode(&sample); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate required fields
	if sample.CollectionId == 0 {
		http.Error(w, "collection_id is required", http.StatusBadRequest)
		return
	}
	if sample.Note == nil {
		note := ""
		sample.Note = &note
	}

	// Set the created_at timestamp
	sample.CreatedAt = time.Now().Unix()

	// Create the database transaction
	tx, err := DB.Begin()
	if err != nil {
		http.Error(w, "error inserting to database", http.StatusInternalServerError)
		return
	}

	// Attempt insert of sample
	result, err := tx.Exec("INSERT INTO samples (collection_id, created_at, note) VALUES (?, ?, ?)", sample.CollectionId, sample.CreatedAt, *sample.Note)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			http.Error(w, "error inserting to database", http.StatusInternalServerError)
			log.Panic(err, rollbackErr)
			return
		}
		http.Error(w, "error inserting to database", http.StatusInternalServerError)
		return
	}
	// Get the inserted ID
	sample_id, err := result.LastInsertId()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			http.Error(w, "error inserting to database", http.StatusInternalServerError)
			log.Panic(err, rollbackErr)
			return
		}
		http.Error(w, "error inserting to database", http.StatusInternalServerError)
		return
	}
	_id := int(sample_id)
	sample.SampleId = &_id

	if len(sample.Values) != 0 {
		// Generate insert query and array of values
		query := "INSERT INTO sample_attribute_values (sample_id, attribute_id, value) VALUES "
		vals := []interface{}{}

		for _, row := range sample.Values {
			query += "(?, ?, ?),"
			vals = append(vals, *sample.SampleId, row.AttributeId, row.Value)
		}
		query = strings.TrimSuffix(query, ",")

		// Attempt insert of values
		_, err = tx.Exec(query, vals...)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				http.Error(w, "error inserting to database", http.StatusInternalServerError)
				log.Panic(err, rollbackErr)
				return
			}
			if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
				http.Error(w, "sample already exists", http.StatusBadRequest)
				return
			}
			http.Error(w, "error inserting to database", http.StatusInternalServerError)
			return
		}
	}

	// Commit the insert
	if err = tx.Commit(); err != nil {
		http.Error(w, "error inserting to database", http.StatusInternalServerError)
		log.Panic(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{\"id\":  %d}", sample_id)
}

// Sample represents a sample entry in the database
type InsertSampleBody struct {
	SampleId     *int          `json:"sample_id,omitempty"`
	CollectionId int           `json:"collection_id,omitempty"`
	CreatedAt    int64         `json:"created_at,omitempty"` // UNIX timestamp
	Note         *string       `json:"note,omitempty"`
	Values       []SampleValue `json:"values"`
}

type SampleValue struct {
	AttributeId int    `json:"attribute_id"`
	Value       string `json:"value"`
}
