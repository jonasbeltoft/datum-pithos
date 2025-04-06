package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
	// Enable foreign keys
	_, err = DB.Exec("PRAGMA foreign_keys = ON")
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

		r.Get("/collections", fetchCollectionsHandler)
		r.Post("/collections", insertCollectionHandler)

		r.Post("/attributes", insertAttributesHandler)
		r.Delete("/attributes", deleteAttributesHandler)

		r.Get("/samples", fetchSamplesHandler)
		r.Post("/samples", insertSampleHandler)
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
