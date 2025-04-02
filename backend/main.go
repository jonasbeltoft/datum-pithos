package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

var db_files string = "/db/"

// DB is a global variable for the SQLite database connection
var DB *sql.DB

const environment = "docker"

func main() {
	if environment == "test" {
		db_files = "../db/"
	}

	var err error

	// Init DB

	// Mount the database file or create if not exists
	DB, err = sql.Open("sqlite3", db_files+"data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// Run init.sql to setup schemas if not exists
	file, err := os.Open(db_files + "init.sql")
	if err != nil {
		log.Fatal(err)
	}
	init, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	_, err = DB.Exec(string(init))
	if err != nil {
		log.Fatal(err)
	}

	// // Setup API functions
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Post("/register", registerHandler)
		r.Post("/login", loginHandler)
	})
	// Private route (requires auth token)
	// user := r.Context().Value("user").(User) is available in these methods
	r.Group(func(r chi.Router) {
		r.Use(AuthenticationMiddleware)
		r.Post("/logout", logoutHandler)
		r.Get("/profile", profileHandler)

		r.Get("/units", fetchUnitsHandler)
		r.Post("/units", insertUnitHandler)
	})

	fmt.Println("Starting server on :8000")
	http.ListenAndServe(":8000", r)
}

type User struct {
	Id              int
	Username        string
	HashedPassword  string
	AccessToken     *string // Nonce
	DisplayName     *string
	Role            *string // Typically this wont be available unless specifically fetched
	RoleId          *int
	TokenExpiryDate *int64 // UNIX time
}

/*
Inserts a new unit into the units table

Body:

	{
		name: string
	}
*/
func insertUnitHandler(w http.ResponseWriter, r *http.Request) {
	// Get the name of the unit
	var body InsertUnitBody

	// Read and decode the JSON body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	} else if len(body.Name) > 32 {
		http.Error(w, "name must be at most 32 characters", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO units (name) VALUES (?)"
	result, err := DB.Exec(query, body.Name)
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

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{id:  %d}", id)
}

type InsertUnitBody struct {
	Name string `json:"name"`
}

/*
Gets one or more units from db

Query params:

	id: int

Result:

	[{
		id: int
		name: string
	}]
*/
func fetchUnitsHandler(w http.ResponseWriter, r *http.Request) {
	var units []Unit
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
	result, err := json.Marshal(FetchUnitsBody{Units: units})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(result))
}

type FetchUnitsBody struct {
	Units []Unit `json:"units"`
}
type Unit struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
