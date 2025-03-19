package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	Username       string
	HashedPassword string
	AccessToken    string
	RefreshToken   string
	DisplayName    string
	Role           string
}

// This should be the DATABASE
var users = map[string]User{}
var sessions = map[string]string{}

func main() {
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/profile", profile)

	fmt.Println("Starting server on :8000")
	http.ListenAndServe(":8000", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the username and password from the request body
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	} else if len(username) < 8 || len(password) < 8 {
		http.Error(w, "Username and password must be at least 8 characters", http.StatusBadRequest)
		return
	} else if len(username) > 50 || len(password) > 50 {
		http.Error(w, "Username and password must be at most 50 characters", http.StatusBadRequest)
		return
	}

	// Check if the user already exists
	if _, ok := users[username]; ok {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Failed create user", http.StatusInternalServerError)
		return
	}

	// Create a new user
	users[username] = User{
		Username:       username,
		HashedPassword: hashedPassword,
		DisplayName:    username,
		Role:           r.FormValue("role"),
	}

	fmt.Fprintln(w, "User created successfully")
}

func login(w http.ResponseWriter, r *http.Request) {
	// Log the request
	fmt.Println(r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the username and password from the request body

	decoder := json.NewDecoder(r.Body)
	var credentials LoginBody
	err := decoder.Decode(&credentials)
	if err != nil {
		panic(err)
	}
	fmt.Println(credentials.Password, credentials.Username)

	if credentials.Username == "" || credentials.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Check if the user exists and the password is correct
	user, ok := users[credentials.Username]
	if !ok || !checkPasswordHash(credentials.Password, user.HashedPassword) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate a new access token
	accessToken := generateNonce(32)
	// Generate a new Refresh token
	refreshToken := generateNonce(32)

	// Save the access token to the user
	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	users[user.Username] = user
	sessions[accessToken] = user.Username

	fmt.Fprintln(w, "{ \"accessToken\": \""+user.AccessToken+"\", \"refreshToken\": \""+user.RefreshToken+"\" }")
}

func logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if the user is authorized
	user, err := Authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Clear the access token and Refresh token
	user.AccessToken = ""
	user.RefreshToken = ""
	users[user.Username] = user
	delete(sessions, user.AccessToken)

	fmt.Fprintln(w, "User logged out successfully")
}

func profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if the user is authorized
	user, err := Authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Fprintln(w, "{ \"username\": \""+user.Username+"\", \"displayName\": \""+user.DisplayName+"\", \"role\": \""+user.Role+"\" }")
}

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
