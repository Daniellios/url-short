package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"example.com/url-short/internal/database"
	"example.com/url-short/internal/service"
	"github.com/joho/godotenv"

	"errors"

	"github.com/lib/pq"
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
	godotenv.Load()

	servMux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: servMux,
	}

	apiConfig := apiConfig{
		db: &database.Queries{},
	}

	dbURL := os.Getenv("DB_URL")
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
