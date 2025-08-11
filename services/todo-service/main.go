package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	UserID    *int64    `json:"user_id,omitempty"`
	User      *User     `json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Server struct {
	db *sql.DB
}

func main() {
	db, err := openDatabaseFromEnv()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := migrate(db); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	server := &Server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/todos", server.handleTodos)
	mux.HandleFunc("/api/todos/", server.handleTodos)

	addr := ":8081"
	if v := os.Getenv("SERVICE_PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("todo-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCommonMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

func withCommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic logging
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func openDatabaseFromEnv() (*sql.DB, error) {
	host := getenvDefault("DB_HOST", "127.0.0.1")
	port := getenvDefault("DB_PORT", "3306")
	user := getenvDefault("DB_USER", "root")
	pass := getenvDefault("DB_PASSWORD", "")
	name := getenvDefault("DB_NAME", "app_db")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, pass, host, port, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Wait for db to be reachable
	for i := 0; i < 60; i++ {
		if err := db.Ping(); err == nil {
			return db, nil
		}
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("database not reachable")
}

func migrate(db *sql.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS todos (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        completed BOOLEAN NOT NULL DEFAULT FALSE,
        user_id BIGINT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB;`
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	// Ensure user_id exists when upgrading from previous schema without using IF NOT EXISTS
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'todos' AND COLUMN_NAME = 'user_id'`).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		if _, err := db.Exec("ALTER TABLE todos ADD COLUMN user_id BIGINT NULL"); err != nil {
			// If the column was concurrently added, ignore duplicate column error
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
				return err
			}
		}
	}
	return nil
}

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		if r.URL.Path == "/api/todos" {
			s.listTodos(w, r)
			return
		}
		http.NotFound(w, r)
	case http.MethodPost:
		if r.URL.Path == "/api/todos" {
			s.createTodo(w, r)
			return
		}
		http.NotFound(w, r)
	case http.MethodPut:
		id, ok := parseID(r.URL.Path, "/api/todos/")
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.updateTodo(w, r, id)
	case http.MethodDelete:
		id, ok := parseID(r.URL.Path, "/api/todos/")
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.deleteTodo(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

func parseID(path string, prefix string) (int64, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	idStr := strings.TrimPrefix(path, prefix)
	if idStr == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *Server) listTodos(w http.ResponseWriter, r *http.Request) {
	includeUser := r.URL.Query().Get("include_user") == "true"
	rows, err := s.db.Query("SELECT id, title, completed, user_id, created_at, updated_at FROM todos ORDER BY id DESC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		var userID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &userID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if userID.Valid {
			t.UserID = &userID.Int64
		}
		if includeUser && t.UserID != nil {
			if u, err := fetchUserByID(*t.UserID); err == nil {
				t.User = u
			}
		}
		todos = append(todos, t)
	}
	_ = json.NewEncoder(w).Encode(todos)
}

func (s *Server) createTodo(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title  string `json:"title"`
		UserID *int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(payload.Title) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "title is required"})
		return
	}
	if payload.UserID != nil {
		if err := validateUserExists(*payload.UserID); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	}
	res, err := s.db.Exec("INSERT INTO todos (title, completed, user_id) VALUES (?, false, ?)", payload.Title, payload.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	var t Todo
	row := s.db.QueryRow("SELECT id, title, completed, user_id, created_at, updated_at FROM todos WHERE id = ?", id)
	var userID sql.NullInt64
	if err := row.Scan(&t.ID, &t.Title, &t.Completed, &userID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if userID.Valid {
		t.UserID = &userID.Int64
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

func (s *Server) updateTodo(w http.ResponseWriter, r *http.Request, id int64) {
	var payload struct {
		Title     *string `json:"title"`
		Completed *bool   `json:"completed"`
		UserID    *int64  `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}

	// Fetch current
	var current Todo
	row := s.db.QueryRow("SELECT id, title, completed, user_id, created_at, updated_at FROM todos WHERE id = ?", id)
	var userID sql.NullInt64
	if err := row.Scan(&current.ID, &current.Title, &current.Completed, &userID, &current.CreatedAt, &current.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if userID.Valid {
		current.UserID = &userID.Int64
	}

	if payload.Title != nil {
		current.Title = *payload.Title
	}
	if payload.Completed != nil {
		current.Completed = *payload.Completed
	}
	if payload.UserID != nil {
		if err := validateUserExists(*payload.UserID); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		current.UserID = payload.UserID
	}

	if _, err := s.db.Exec("UPDATE todos SET title = ?, completed = ?, user_id = ? WHERE id = ?", current.Title, current.Completed, current.UserID, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	row = s.db.QueryRow("SELECT id, title, completed, user_id, created_at, updated_at FROM todos WHERE id = ?", id)
	if err := row.Scan(&current.ID, &current.Title, &current.Completed, &userID, &current.CreatedAt, &current.UpdatedAt); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if userID.Valid {
		current.UserID = &userID.Int64
	} else {
		current.UserID = nil
	}
	_ = json.NewEncoder(w).Encode(current)
}

func (s *Server) deleteTodo(w http.ResponseWriter, r *http.Request, id int64) {
	res, err := s.db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func validateUserExists(userID int64) error {
	u, err := fetchUserByID(userID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user not found")
	}
	return nil
}

func fetchUserByID(id int64) (*User, error) {
	// Direct service-to-service call inside Docker network; overridable via env USERS_SERVICE_BASE
	base := getenvDefault("USERS_SERVICE_BASE", "http://users-service:8082")
	endpoint := fmt.Sprintf("%s/api/users/%d", strings.TrimRight(base, "/"), id)
	client := &http.Client{Timeout: 3 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("user not found")
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("users-service error: %s", strings.TrimSpace(string(b)))
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
