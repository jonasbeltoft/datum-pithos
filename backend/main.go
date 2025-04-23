package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	DB, err = sql.Open("sqlite3", "file:"+db_files+"data.db?_foreign_keys=on&cache=shared&busy_timeout=5000&journal_mode=WAL")
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

	baseApirUrl := "/api/v1/"

	// Public routes
	r.Group(func(r chi.Router) {
		r.Use(dbLoggerMiddleware)
		r.Post(baseApirUrl+"login", loginHandler)
	})
	// Private route (requires auth token)
	// user := r.Context().Value("user").(User) is available in these methods
	r.Group(func(r chi.Router) {
		r.Use(AuthenticationMiddleware, dbLoggerMiddleware)
		r.Post(baseApirUrl+"logout", logoutHandler)
		r.Get(baseApirUrl+"auth", authHandler)

		r.Get(baseApirUrl+"units", fetchUnitsHandler)
		r.Post(baseApirUrl+"units", insertUnitHandler)

		r.Get(baseApirUrl+"collections", fetchCollectionsHandler)
		r.Post(baseApirUrl+"collections", insertCollectionHandler)
		r.Delete(baseApirUrl+"collections", deleteCollectionHandler)

		r.Post(baseApirUrl+"attributes", insertAttributesHandler)
		r.Delete(baseApirUrl+"attributes", deleteAttributesHandler)

		r.Get(baseApirUrl+"samples", fetchSamplesHandler)
		r.Post(baseApirUrl+"samples", insertSampleHandler)
		r.Delete(baseApirUrl+"samples", deleteSampleHandler)
		r.Put(baseApirUrl+"samples", updateSampleHandler)
		r.Post(baseApirUrl+"sample-values", insertOrUpdateSampleValueHandler)
	})

	// Admin routes (requires admin role)
	r.Group(func(r chi.Router) {
		r.Use(AuthenticationMiddleware, dbLoggerMiddleware, AdminMiddleware)
		r.Get(baseApirUrl+"logs", fetchLogsHandler)

		r.Get(baseApirUrl+"users", fetchUsersHandler)
		r.Post(baseApirUrl+"users", insertUserHandler)
		r.Delete(baseApirUrl+"users", deleteUserHandler)
		r.Put(baseApirUrl+"users", updateUserHandler)

		r.Get(baseApirUrl+"roles", fetchRolesHandler)
	})

	// Init admin user if not exists
	if err := initAdminUser(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting server on :8000")
	http.ListenAndServe(":8000", r)
}

func initAdminUser() error {
	// Check if the admin user exists
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Admin user already exists, no need to create it
		return nil
	}

	// Create the admin user
	// Read the 'auth_init.txt' file to get the username and password
	file, err := os.Open(db_files + "auth_init.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file line by line
	var username, password string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "username=") {
			username = strings.TrimPrefix(line, "username=")
		} else if strings.HasPrefix(line, "password=") {
			password = strings.TrimPrefix(line, "password=")
		}
	}

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	role_id := new(int)
	*role_id = 1

	// Create a new user
	user := User{
		Username:       username,
		HashedPassword: hashedPassword,
		DisplayName:    &username,
		RoleId:         role_id,
	}

	user, err = createUser(user)
	if err != nil {
		return err
	}

	fmt.Println("Admin user created with username:", user.Username)

	return nil
}

type User struct {
	Id              int     `json:"id"`
	Username        string  `json:"username"`
	HashedPassword  string  `json:"-"` // Don't send this to the client
	AccessToken     *string `json:"-"` // Nonce
	DisplayName     *string `json:"display_name"`
	Role            *string `json:"-"` // Typically this wont be available unless specifically fetched
	RoleId          *int    `json:"role_id"`
	TokenExpiryDate *int64  `json:"-"` // UNIX time
}
