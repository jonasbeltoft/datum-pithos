package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

var db_files string = "./db/"

// DB is a global variable for the SQLite database connection
var DB *sql.DB

func main() {
	var err error

	// Init DB
	// If ./db directory does not exist, create it
	if _, err := os.Stat(db_files); os.IsNotExist(err) {
		err = os.MkdirAll(db_files, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Mount the database file or create if not exists
	DB, err = sql.Open("sqlite3", "file:"+db_files+"data.db?_foreign_keys=on&cache=shared&busy_timeout=5000&journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// Run init to setup schemas if not exists

	_, err = DB.Exec(db_init)
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

	// read input args for port
	port := ":8000"
	args := os.Args
	if len(args) > 1 {
		if strings.HasPrefix(args[1], "--port=") {
			_port := strings.TrimPrefix(args[1], "--port=")
			if _port != "" {
				port = ":" + _port
			}
		} else {
			fmt.Println("Invalid argument. Use --port=<port_number>")
			return
		}
	}

	fmt.Println("Starting server on port", port)
	http.ListenAndServe(port, r)
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

	// Read the env to get the username and password
	// username := os.Getenv("USERNAME")
	// if username == "" {
	// 	return fmt.Errorf("USERNAME environment variable not set")
	// }

	// password := os.Getenv("PASSWORD")
	// if password == "" {
	// 	return fmt.Errorf("PASSWORD environment variable not set")
	// }
	username := "admin"
	password := "admin"

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	role_id := 1

	// Create a new user
	user := User{
		Username:       username,
		HashedPassword: hashedPassword,
		DisplayName:    &username,
		RoleId:         &role_id,
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

var db_init = `
-- Create table: units
CREATE TABLE IF NOT EXISTS units (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT unique_name UNIQUE (name)
);

-- Create table: collections
CREATE TABLE IF NOT EXISTS collections (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    CONSTRAINT unique_name UNIQUE (name)
);

-- Create table: samples
CREATE TABLE IF NOT EXISTS samples (
    id INTEGER PRIMARY KEY,
    collection_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL, -- Stores UNIX time at INSERT
    note TEXT,
    CONSTRAINT fk_collection FOREIGN KEY (collection_id) REFERENCES collections (id) ON DELETE CASCADE
);

-- Create table: sample_attributes
CREATE TABLE IF NOT EXISTS sample_attributes (
    id INTEGER PRIMARY KEY,
    collection_id INTEGER NOT NULL,
    unit_id INTEGER, -- Nullable
    name TEXT NOT NULL,
    CONSTRAINT unique_name UNIQUE (collection_id, name),
    CONSTRAINT fk_collection FOREIGN KEY (collection_id) REFERENCES collections (id) ON DELETE CASCADE,
    CONSTRAINT fk_unit FOREIGN KEY (unit_id) REFERENCES units (id)
);

-- Create table: sample_attribute_values
CREATE TABLE IF NOT EXISTS sample_attribute_values (
    sample_id INTEGER NOT NULL,
    attribute_id INTEGER NOT NULL,
    value TEXT,
    PRIMARY KEY (sample_id, attribute_id),
    CONSTRAINT fk_sample FOREIGN KEY (sample_id) REFERENCES samples (id) ON DELETE CASCADE,
    CONSTRAINT fk_attribute FOREIGN KEY (attribute_id) REFERENCES sample_attributes (id) ON DELETE CASCADE
);

-- Create table: roles
CREATE TABLE IF NOT EXISTS roles (id INTEGER PRIMARY KEY, name TEXT NOT NULL);

-- Create table: users
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role_id INTEGER, -- Nullable, references roles table
    display_name TEXT,
    access_token TEXT,
    token_expiry_date INTEGER, -- UNIX time in seconds
    CONSTRAINT unique_username UNIQUE (username),
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES roles (id)
);

-- Create table: logs
CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY,
    created_at INTEGER NOT NULL, -- Stores UNIX time
    instance_user INTEGER, -- References users
    crud_action TEXT NOT NULL, -- Action performed
    request_url TEXT,
    request_body TEXT,
    response_code INTEGER
);

-- Initialize roles table only if empty
INSERT INTO
    roles
SELECT
    *
FROM
    (
        VALUES
            (1, 'admin'),
            (2, 'Lab Technician')
    ) source_data
WHERE
    NOT EXISTS (
        SELECT
            NULL
        FROM
            roles
    );`
