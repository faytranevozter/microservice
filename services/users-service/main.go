package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
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
	mux.HandleFunc("/api/users", server.handleUsers)
	mux.HandleFunc("/api/users/", server.handleUserByID)

	addr := ":8082"
	if v := os.Getenv("SERVICE_PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("users-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCommonMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

func withCommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	for i := 0; i < 60; i++ {
		if err := db.Ping(); err == nil {
			return db, nil
		}
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("database not reachable")
}

func migrate(db *sql.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS users (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        email VARCHAR(255) NOT NULL UNIQUE,
        name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ) ENGINE=InnoDB;`
	_, err := db.Exec(schema)
	return err
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		s.listUsers(w, r)
	case http.MethodPost:
		s.createUser(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	id, ok := parseID(r.URL.Path, "/api/users/")
	if !ok {
		http.NotFound(w, r)
		return
	}
	var u User
	row := s.db.QueryRow("SELECT id, email, name, created_at FROM users WHERE id = ?", id)
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(u)
}

func parseID(path string, prefix string) (int64, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	idStr := strings.TrimPrefix(path, prefix)
	if idStr == "" {
		return 0, false
	}
	var id int64
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id, email, name, created_at FROM users ORDER BY id DESC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		users = append(users, u)
	}
	_ = json.NewEncoder(w).Encode(users)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(payload.Email) == "" || strings.TrimSpace(payload.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "email and name are required"})
		return
	}
	res, err := s.db.Exec("INSERT INTO users (email, name) VALUES (?, ?)", payload.Email, payload.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	var u User
	row := s.db.QueryRow("SELECT id, email, name, created_at FROM users WHERE id = ?", id)
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(u)
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
