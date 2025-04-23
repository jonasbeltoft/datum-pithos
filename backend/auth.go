package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AdminMiddleware checks if the user is an admin (role_id = 1)
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)

		if user.RoleId == nil || *user.RoleId != 1 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Proceed to the next handler if authorized
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware checks for an Authorization header
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		accessToken = strings.TrimPrefix(accessToken, "Bearer ")

		user, err := readUserByAccesToken(accessToken, false, true)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check for token timeout
		if *user.TokenExpiryDate < time.Now().Unix() {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)

		// Proceed to the next handler if authorized
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

/*
Update a user by their ID

Params:

	user_id: int
	display_name: string
	role_id: int
*/
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	displayName := r.FormValue("display_name")
	roleId := r.FormValue("role_id")
	_userId := r.FormValue("user_id")

	var err error

	var userId int
	if _userId == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	} else {
		userId, err = strconv.Atoi(_userId)
		if err != nil {
			http.Error(w, "Invalid User ID", http.StatusBadRequest)
			return
		}
	}

	if displayName == "" && roleId == "" {
		http.Error(w, "At least one of display_name or role_id is required", http.StatusBadRequest)
		return
	}
	if len(displayName) > 50 {
		http.Error(w, "Display name must be at most 50 characters", http.StatusBadRequest)
		return
	} else if len(displayName) < 3 {
		http.Error(w, "Display name must be at least 3 characters", http.StatusBadRequest)
		return
	}

	var roleIdInt *int
	if roleId != "" {
		rid, err := strconv.Atoi(roleId)
		if err != nil {
			http.Error(w, "Invalid Role ID", http.StatusBadRequest)
			return
		}
		roleIdInt = &rid
	}

	user := User{
		Id: userId,
	}
	if roleId != "" {
		user.RoleId = roleIdInt
	}
	if displayName != "" {
		user.DisplayName = &displayName
	}

	err = updateUserBasic(user)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "User updated successfully")
}

/*
Get all roles
*/
func fetchRolesHandler(w http.ResponseWriter, r *http.Request) {
	var roles []Role = []Role{}
	var err error

	query := "SELECT id, name FROM roles;"
	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var role Role
		err := rows.Scan(&role.Id, &role.Name)
		if err != nil {
			http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
			return
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(roles)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(result))
}

type Role struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

/*
Delete a user by their ID

Params:

	user_id: int
*/
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	_userId := r.FormValue("user_id")
	if _userId == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(_userId)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	deleteUser(userId)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "User deleted successfully")
}

/*
Fetch all users
*/
func fetchUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User = []User{}
	var err error

	query := "SELECT id, username, role_id, display_name FROM users;"
	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Username, &user.RoleId, &user.DisplayName)
		if err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(users)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(result))
}

/*
Params:

	username: string
	password: string
	role_id: int
*/
func insertUserHandler(w http.ResponseWriter, r *http.Request) {
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

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Failed create user", http.StatusInternalServerError)
		return
	}

	// Convert RoleId to int
	r_id, err := strconv.Atoi(r.FormValue("role_id"))
	if err != nil {
		http.Error(w, "Failed create user", http.StatusBadRequest)
		return
	}

	// Create a new user
	user := User{
		Username:       username,
		HashedPassword: hashedPassword,
		DisplayName:    &username,
		RoleId:         &r_id,
	}

	user, err = createUser(user)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			http.Error(w, "Username already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed create user", http.StatusInternalServerError)
		return
	}

	outUser := User{
		Id:          user.Id,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		RoleId:      user.RoleId,
	}

	json.NewEncoder(w).Encode(outUser)
}

/*
Body:

	{
		username: string
		password: string
	}
*/
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Get the username and password from the request body
	decoder := json.NewDecoder(r.Body)
	var credentials LoginBody
	err := decoder.Decode(&credentials)
	if err != nil {
		panic(err)
	}

	if credentials.Username == "" || credentials.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Check if the user exists and the password is correct
	user, err := readUserByUsername(credentials.Username, true, false)
	if err != nil || !checkPasswordHash(credentials.Password, user.HashedPassword) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate a new access token
	accessToken := generateNonce(32)
	expiryTime := time.Now().Add(12 * time.Hour).Unix()

	// Save the access token to the user
	user.AccessToken = &accessToken
	user.TokenExpiryDate = &expiryTime

	err = updateUserAuth(user.Id, &accessToken, &expiryTime)
	if err != nil {
		http.Error(w, "An error occured when creating session", http.StatusUnauthorized)
		return
	}

	fmt.Fprintln(w, "{ \"accessToken\": \""+*user.AccessToken+"\", \"tokenExpiryDate\": \""+strconv.Itoa(int(*user.TokenExpiryDate))+"\" }")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	// Clear the access token and Refresh token
	err := updateUserAuth(user.Id, nil, nil)
	if err != nil {
		http.Error(w, "An error occured when creating session", http.StatusUnauthorized)
		return
	}

	fmt.Fprintln(w, "User logged out successfully")
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	fmt.Fprintln(w, "{ \"username\": \""+user.Username+"\", \"displayName\": \""+*user.DisplayName+"\", \"role\": \""+*user.Role+"\" }")
}

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateNonce(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

// Uses RoleId
func createUser(user User) (User, error) {
	// Prepare the INSERT statement with a subquery
	query := `
		INSERT INTO users (username, password_hash, role_id, display_name) 
		VALUES (?, ?, ?, ?);
	`

	result, err := DB.Exec(query, user.Username, user.HashedPassword, user.RoleId, user.DisplayName)
	if err != nil {
		return User{}, fmt.Errorf("createUserExec: %v", err)
	}

	// Get the new user's generated ID for the client.
	id, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("createUserId: %v", err)
	}

	user.Id = int(id)

	return user, nil
}

// Reads a user by their ID
func readUser(userID int, include_password bool, include_session bool) (User, error) {
	// Prepare the SELECT query
	q := `
		SELECT id, username, role_id, display_name`
	q_pass := `, password_hash`
	q_auth := `, access_token, token_expiry_date`
	q_end := `
		FROM users 
		WHERE id = ?;
	`

	if include_password {
		q += q_pass
	}
	if include_session {
		q += q_auth
	}
	q += q_end

	// Create a User object to store the result
	var user User
	includes := []interface{}{&user.Id, &user.Username, &user.RoleId, &user.DisplayName}
	if include_password {
		includes = append(includes, &user.HashedPassword)
	}
	if include_session {
		includes = append(includes, &user.AccessToken, &user.TokenExpiryDate)
	}
	// Execute the query and scan the result into the User object
	err := DB.QueryRow(q, userID).Scan(includes...)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("readUser: no user found with ID %d", userID)
		}
		return User{}, fmt.Errorf("readUser: %v", err)
	}

	// Execute the query and scan the result into the User object
	err = DB.QueryRow("SELECT name FROM roles WHERE id = ?;", user.RoleId).Scan(&user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("readRole: no role found with ID %d", user.RoleId)
		}
		return User{}, fmt.Errorf("readRole: %v", err)
	}

	return user, nil
}

// Reads a user by their ID
func readUserByAccesToken(access_token string, include_password bool, include_session bool) (User, error) {
	// Prepare the SELECT query
	q := `
		SELECT id, username, role_id, display_name`
	q_pass := `, password_hash`
	q_auth := `, access_token, token_expiry_date`
	q_end := `
		FROM users 
		WHERE access_token = ?;
	`

	if include_password {
		q += q_pass
	}
	if include_session {
		q += q_auth
	}
	q += q_end

	// Create a User object to store the result
	var user User
	includes := []interface{}{&user.Id, &user.Username, &user.RoleId, &user.DisplayName}
	if include_password {
		includes = append(includes, &user.HashedPassword)
	}
	if include_session {
		includes = append(includes, &user.AccessToken, &user.TokenExpiryDate)
	}
	// Execute the query and scan the result into the User object
	err := DB.QueryRow(q, access_token).Scan(includes...)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("readUserByAccesToken: no user found with token %s", access_token)
		}
		return User{}, fmt.Errorf("readUserByAccesToken: %v", err)
	}

	// Execute the query and scan the result into the User object
	err = DB.QueryRow("SELECT name FROM roles WHERE id = ?;", user.RoleId).Scan(&user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("readRole: no role found with token %d", user.RoleId)
		}
		return User{}, fmt.Errorf("readRole: %v", err)
	}

	return user, nil
}

// Reads a user by their ID
func readUserByUsername(username string, include_password bool, include_session bool) (User, error) {
	// Prepare the SELECT query
	q := `
		SELECT id, username, role_id, display_name`
	q_pass := `, password_hash`
	q_auth := `, access_token, token_expiry_date`
	q_end := `
		FROM users 
		WHERE username = ?;
	`

	if include_password {
		q += q_pass
	}
	if include_session {
		q += q_auth
	}
	q += q_end

	// Create a User object to store the result
	var user User
	includes := []interface{}{&user.Id, &user.Username, &user.RoleId, &user.DisplayName}
	if include_password {
		includes = append(includes, &user.HashedPassword)
	}
	if include_session {
		includes = append(includes, &user.AccessToken, &user.TokenExpiryDate)
	}
	// Execute the query and scan the result into the User object
	err := DB.QueryRow(q, username).Scan(includes...)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("readUserByUsername: no user found with username %s", username)
		}
		return User{}, fmt.Errorf("readUser: %v", err)
	}

	// Execute the query and scan the result into the User object
	err = DB.QueryRow("SELECT name FROM roles WHERE id = ?;", user.RoleId).Scan(&user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("readRole: no role found with ID %d", user.RoleId)
		}
		return User{}, fmt.Errorf("readRole: %v", err)
	}

	return user, nil
}

// Deletes a user by their ID
func deleteUser(userID int) error {
	// Prepare the DELETE query
	query := `
		DELETE FROM users 
		WHERE id = ?;
	`

	// Execute the DELETE query
	result, err := DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("deleteUser: %v", err)
	}

	// Check if a row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleteUserRowsAffected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("deleteUser: no user found with ID %d", userID)
	}

	return nil
}

// Updates the superficial data of a user (e.g., display_name and role_id)
func updateUserBasic(user User) error {
	// Prepare the UPDATE query
	query := []string{}
	startQ := "UPDATE users SET "
	var args []any = []any{}

	if user.DisplayName != nil {
		query = append(query, "display_name = ?")
		args = append(args, *user.DisplayName)
	}

	if user.RoleId != nil {
		query = append(query, "role_id = ?")
		args = append(args, *user.RoleId)
	}

	if len(args) == 0 {
		return fmt.Errorf("updateUserBasic: no fields to update")
	}

	startQ += strings.Join(query, ", ") + "WHERE id = ?;"
	args = append(args, user.Id)

	// Execute the UPDATE query
	result, err := DB.Exec(startQ, args...)
	if err != nil {
		return fmt.Errorf("updateUserBasic: %v", err)
	}

	// Check if a row was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("updateUserBasicRowsAffected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("updateUserBasic: no user found with ID %d", user.Id)
	}

	return nil
}

// Updates the authentication data of a user (e.g. access_token, and token_expiry_date)
func updateUserAuth(id int, access_token *string, token_expiry_date *int64) error {
	// Prepare the UPDATE query
	query := `
		UPDATE users 
		SET access_token = ?, token_expiry_date = ? 
		WHERE id = ?;
	`

	// Execute the UPDATE query
	result, err := DB.Exec(query, access_token, token_expiry_date, id)
	if err != nil {
		return fmt.Errorf("updateUserAuth: %v", err)
	}

	// Check if a row was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("updateUserAuthRowsAffected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("updateUserAuth: no user found with ID %d", id)
	}

	return nil
}
