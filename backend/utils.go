package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

func Authenticate(r *http.Request) (User, error) {
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		return User{}, errors.New("authorization header is missing")
	}

	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	username, ok := sessions[accessToken]
	if !ok {
		fmt.Println("Session not found")
		return User{}, ErrAuth
	}

	user, ok := users[username]
	if !ok {
		fmt.Println("User not found")
		return User{}, ErrAuth
	}

	// check access token against users current token
	if user.AccessToken != accessToken {
		fmt.Println("Access token mismatch")
		delete(sessions, accessToken)
		return User{}, ErrAuth
	}

	return user, nil
}
