package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5/middleware"
)

func dbLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_user := r.Context().Value("user")
		var user User
		if _user == nil {
			user = User{Id: -1}
		} else {
			user = _user.(User)
		}

		// Save the raw body bytes before decoding
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("failed to read request body", err.Error())
		} else if len(bodyBytes) == 0 {
			bodyBytes = []byte("")
		} else {
			// Restore the body so the next handler can use it
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Wrap the response writer to capture the status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		url := r.URL.RequestURI()

		// Hide password in logs
		if r.URL.Path == "/register" {
			params := strings.Split(url, "&")
			for i, param := range params {
				if strings.Contains(param, "password") {
					params[i] = "password=****"
				}
			}
			url = strings.Join(params, "&")
		}
		if r.URL.Path == "/login" {
			var body LoginBody
			err = json.Unmarshal(bodyBytes, &body)
			if err != nil {
				log.Println("failed to unmarshal request body", err.Error())
			} else {
				// Hide password in logs
				body.Password = "****"
				bodyBytes, err = json.Marshal(body)
				if err != nil {
					log.Println("failed to marshal request body", err.Error())
				}
			}
		}

		// Remove all whitespace from the body
		var b strings.Builder
		b.Grow(len(string(bodyBytes)))
		for _, ch := range string(bodyBytes) {
			if !unicode.IsSpace(ch) {
				b.WriteRune(ch)
			}
		}
		bodyString := b.String()

		action_log := Log{
			CreatedAt:    time.Now().Unix(),
			InstanceUser: user.Id,
			CRUDAction:   r.Method,
			RequestUrl:   r.RemoteAddr + " " + url,
			RequestBody:  bodyString,
			ResponseCode: ww.Status(),
		}
		if err := insertLog(action_log); err != nil {
			log.Println("failed to log request: ", err.Error())
		}
	})
}

func fetchLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	_userId := r.FormValue("user_id")
	httpMethod := strings.ToLower(r.FormValue("http_method"))

	// Validate & build filters
	var filters []string
	var args []interface{}

	if _userId != "" {
		userId, err := strconv.Atoi(_userId)
		if err != nil || userId <= 0 {
			http.Error(w, "user_id must be a positive int", http.StatusBadRequest)
			return
		}
		filters = append(filters, "instance_user = ?")
		args = append(args, userId)
	}

	if httpMethod != "" {
		allowedMethods := map[string]bool{
			"get": true, "post": true, "put": true, "update": true, "delete": true,
		}
		if !allowedMethods[httpMethod] {
			http.Error(w, "http_method must be one of: get, post, put, update, delete", http.StatusBadRequest)
			return
		}
		filters = append(filters, "LOWER(crud_action) = ?")
		args = append(args, httpMethod)
	}

	query := `
		SELECT id, created_at, instance_user, crud_action, request_url, request_body, response_code
		FROM logs
	`
	if len(filters) > 0 {
		query += "WHERE " + strings.Join(filters, " AND ") + " "
	}
	query += "ORDER BY created_at DESC"

	// Execute query
	rows, err := DB.Query(query, args...)
	if err != nil {
		http.Error(w, "error querying logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []Log = make([]Log, 0)
	for rows.Next() {
		var log Log
		if err := rows.Scan(&log.ID, &log.CreatedAt, &log.InstanceUser, &log.CRUDAction, &log.RequestUrl, &log.RequestBody, &log.ResponseCode); err != nil {
			http.Error(w, "error scanning logs", http.StatusInternalServerError)
			return
		}
		logs = append(logs, log)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

type Log struct {
	ID           int    `json:"id"`
	CreatedAt    int64  `json:"created_at"`
	InstanceUser int    `json:"instance_user"`
	CRUDAction   string `json:"crud_action"`
	RequestUrl   string `json:"request_url"`
	RequestBody  string `json:"request_body"`
	ResponseCode int    `json:"response_code"`
}

func insertLog(log Log) error {
	// Insert log into database
	query := `INSERT INTO logs (created_at, instance_user, crud_action, request_url, request_body, response_code) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(query, log.CreatedAt, log.InstanceUser, log.CRUDAction, log.RequestUrl, log.RequestBody, log.ResponseCode)
	if err != nil {
		return fmt.Errorf("failed to insert log: %w", err)
	}
	return nil
}
