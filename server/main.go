package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"example.com/url-short/internal/database"
	"example.com/url-short/internal/service"
	"github.com/joho/godotenv"

	"github.com/lib/pq"
	"github.com/rs/cors"
)

type apiConfig struct {
	db *database.Queries
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func (apiConfig *apiConfig) createShortURL(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		OriginalURL string `json:"original_url"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		service.RespondWithError(w, 400, "Error decoding parameters")
		return
	}

	originalURL := params.OriginalURL

	if originalURL == "" {
		service.RespondWithError(w, 400, "URL is required")
		return
	}

	shortURL, err := service.ShortenURL(7)
	if err != nil {
		log.Printf("generate short URL failed: %v", err)
		service.RespondWithError(w, 500, "Error generating short URL")
		return
	}

	createdURLData, err := apiConfig.db.CreateShortURL(r.Context(), database.CreateShortURLParams{
		OriginalUrl:  originalURL,
		ShortUrlCode: shortURL,
	})

	if err != nil {
		if isUniqueViolation(err) {
			for i := 0; i < 3; i++ {
				shortURL, err := service.ShortenURL(7)
				if err != nil {
					continue
				}
				createdURLData, err = apiConfig.db.CreateShortURL(r.Context(), database.CreateShortURLParams{
					OriginalUrl:  originalURL,
					ShortUrlCode: shortURL,
				})
				if err != nil {
					continue
				}
				break
			}
		}
		log.Printf("create short URL failed: %v", err)
		service.RespondWithError(w, 500, "Error creating short URL")
		return
	}

	shortURL = service.FormatShortURL(createdURLData.ShortUrlCode)
	service.RespondWithJSON(w, http.StatusCreated, map[string]string{"short_url": shortURL})
}

func (apiConfig *apiConfig) redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("short_url")

	if shortURL == "" || len(shortURL) != 7 {
		service.RespondWithError(w, 400, "Invalid short URL")
		return
	}

	originalURLData, err := apiConfig.db.GetShortURL(r.Context(), shortURL)

	if err != nil {
		log.Printf("get short URL failed: %v", err)
		service.RespondWithError(w, 404, "Short URL not found")
		return
	}

	w.Header().Set("Location", originalURLData.OriginalUrl)
	w.WriteHeader(http.StatusMovedPermanently)
}

func main() {
	_ = godotenv.Load("server/.env")
	_ = godotenv.Load(".env")

	if service.IsLocalDev() && strings.TrimSpace(os.Getenv("DOMAIN")) == "" {
		log.Printf("local dev: short links use http://localhost:8080/{code} (set DOMAIN to override)")
	}

	servMux := http.NewServeMux()
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}).Handler(servMux)
	server := http.Server{
		Addr:    ":8080",
		Handler: corsHandler,
	}

	apiConfig := apiConfig{
		db: &database.Queries{},
	}

	dbURL, err := resolveDBURL()
	if err != nil {
		log.Fatal(err)
	}
	dbQueries, db, err := setupDB(dbURL)
	if err != nil {
		log.Fatalf("Error setting up database: %s", err)
	}

	defer db.Close()
	apiConfig.db = dbQueries

	servMux.HandleFunc("POST /api/urls", apiConfig.createShortURL)

	servMux.HandleFunc("GET /{short_url}", apiConfig.redirectToOriginalURL)

	servMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		service.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	server.ListenAndServe()
}

func setupDB(dbURL string) (*database.Queries, *sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, nil, err
	}
	dbQueries := database.New(db)
	return dbQueries, db, nil
}

func resolveDBURL() (string, error) {
	if v := strings.TrimSpace(os.Getenv("DB_URL")); v != "" {
		return v, nil
	}
	if service.IsProduction() {
		return "", fmt.Errorf("DB_URL is required when APP_ENV=production or DOCKER=1")
	}
	pass := strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD"))
	if pass == "" {
		pass = "admin"
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("urlshort", pass),
		Host:   "localhost:5432",
		Path:   "/urlshort",
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	log.Printf("local dev: DB_URL not set; using %s (set DB_URL or POSTGRES_PASSWORD to override)", u.Redacted())
	return u.String(), nil
}
