package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"example.com/url-short/internal/url-short/database"
	"example.com/url-short/internal/url-short/service"
	"example.com/url-short/internal/utils"
	"github.com/joho/godotenv"

	"github.com/rs/cors"
)

type apiConfig struct {
	db                    *database.Queries
	localDevDefaultDomain string
}

func (apiConfig *apiConfig) createShortURL(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		OriginalURL string `json:"original_url"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, 400, "Error decoding parameters")
		return
	}

	originalURL := params.OriginalURL

	if originalURL == "" {
		utils.RespondWithError(w, 400, "URL is required")
		return
	}

	shortURL, err := service.ShortenURL(7)
	if err != nil {
		log.Printf("generate short URL failed: %v", err)
		utils.RespondWithError(w, 500, "Error generating short URL")
		return
	}

	createdURLData, err := apiConfig.db.CreateShortURL(r.Context(), database.CreateShortURLParams{
		OriginalUrl:  originalURL,
		ShortUrlCode: shortURL,
	})

	if err != nil {
		if utils.IsUniqueViolation(err) {
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
		utils.RespondWithError(w, 500, "Error creating short URL")
		return
	}

	shortURL = utils.FormatShortURL(createdURLData.ShortUrlCode, apiConfig.localDevDefaultDomain)
	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"short_url": shortURL})
}

func (apiConfig *apiConfig) redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("short_url")

	if shortURL == "" || len(shortURL) != 7 {
		utils.RespondWithError(w, 400, "Invalid short URL")
		return
	}

	originalURLData, err := apiConfig.db.GetShortURL(r.Context(), shortURL)

	if err != nil {
		log.Printf("get short URL failed: %v", err)
		utils.RespondWithError(w, 404, "Short URL not found")
		return
	}

	w.Header().Set("Location", originalURLData.OriginalUrl)
	w.WriteHeader(http.StatusMovedPermanently)
}

func main() {
	_ = godotenv.Load("server/.env")
	_ = godotenv.Load(".env")

	if utils.IsLocalDev() && strings.TrimSpace(os.Getenv("DOMAIN")) == "" {
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
		db:                    &database.Queries{},
		localDevDefaultDomain: "http://localhost:8080",
	}

	dbURL, err := utils.ResolveDBURL("/urlshort", "urlshort")
	if err != nil {
		log.Fatal(err)
	}

	dbQueries, db, err := utils.SetupDB(dbURL, func(db *sql.DB) *database.Queries {
		return database.New(db)
	})

	if err != nil {
		log.Fatalf("Error setting up database: %s", err)
	}

	defer db.Close()
	apiConfig.db = dbQueries

	servMux.HandleFunc("POST /api/url-short/urls", apiConfig.createShortURL)

	servMux.HandleFunc("GET /api/url-short/{short_url}", apiConfig.redirectToOriginalURL)

	servMux.HandleFunc("GET /api/url-short/healthz", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	server.ListenAndServe()
}
