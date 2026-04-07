package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/lib/pq"
)

func FormatShortURL(shortURL string, localDevDefaultDomain string) string {
	domain := strings.TrimRight(strings.TrimSpace(os.Getenv("DOMAIN")), "/")
	if domain != "" {
		return domain + "/" + shortURL
	}
	if IsLocalDev() {
		return localDevDefaultDomain + "/" + shortURL
	}
	return "/" + shortURL
}

func SetupDB[Q any](dbURL string, newQueries func(*sql.DB) Q) (Q, *sql.DB, error) {
	var zero Q
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return zero, nil, err
	}
	return newQueries(db), db, nil
}

func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func ResolveDBURL(dbPath string, dbUser string) (string, error) {
	if v := strings.TrimSpace(os.Getenv("DB_URL")); v != "" {
		return v, nil
	}
	if IsLocalDev() {
		return "postgres://postgres:admin@localhost:5432" + dbPath + "?sslmode=disable", nil
	}
	if IsProduction() {
		return "", fmt.Errorf("DB_URL is required when APP_ENV=production or DOCKER=1")
	}
	pass := strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD"))
	if pass == "" {
		pass = "admin"
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbUser, pass),
		Host:   "localhost:5432",
		Path:   dbPath,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	log.Printf("local dev: DB_URL not set; using %s (set DB_URL or POSTGRES_PASSWORD to override)", u.Redacted())
	return u.String(), nil
}
