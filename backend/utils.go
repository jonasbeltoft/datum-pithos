package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

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

var ErrAuth = errors.New("Unauthorized")

func Authorize(r *http.Request) (User, error) {
	username := r.FormValue("username")
	user, ok := users[username]
	if !ok {
		fmt.Println("User not found")
		return user, ErrAuth
	}

	// Check the session token
	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" || sessionToken.Value != user.SessionToken {
		fmt.Println("Invalid session token")
		return user, ErrAuth
	}

	// Check the CSRF token
	csrfToken := r.Header.Get("X-CSRF-Token")
	if csrfToken == "" || csrfToken != user.CSRFToken {
		fmt.Println("Invalid CSRF token")
		return user, ErrAuth
	}

	return user, nil
}
