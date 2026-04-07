package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/url-short/internal/pastebin/database"
	"example.com/url-short/internal/utils"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type apiConfig struct {
	db                    *database.Queries
	localDevDefaultDomain string
}

func (apiConfig *apiConfig) createPastebin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Content string `json:"content"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, 400, "Error decoding parameters")
		return
	}

	content := params.Content
	log.Printf("content: %s", content)

	createdPasteData, err := apiConfig.db.CreatePaste(r.Context(), content)
	if err != nil {
		log.Printf("create paste failed: %v", err)
		utils.RespondWithError(w, 500, "Error creating paste")
		return
	}
	log.Printf("created paste data: %+v", createdPasteData)
	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"id":         strconv.FormatInt(createdPasteData.ID, 10),
		"content":    content,
		"created_at": createdPasteData.CreatedAt.Time.Format(time.RFC3339),
		"updated_at": createdPasteData.UpdatedAt.Time.Format(time.RFC3339),
	})
}

func (apiConfig *apiConfig) getPastebin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.RespondWithError(w, 400, "ID is required")
		return
	}
	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		utils.RespondWithError(w, 400, "Invalid ID")
		return
	}
	pasteData, err := apiConfig.db.GetPaste(r.Context(), intId)
	if err != nil {
		log.Printf("get paste failed: %v", err)
		utils.RespondWithError(w, 500, "Error getting paste")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"id":         strconv.FormatInt(pasteData.ID, 10),
		"content":    pasteData.Content,
		"created_at": pasteData.CreatedAt.Time.Format(time.RFC3339),
		"updated_at": pasteData.UpdatedAt.Time.Format(time.RFC3339),
	})

}

func (apiConfig *apiConfig) getAllPastes(w http.ResponseWriter, r *http.Request) {
	pasteData, err := apiConfig.db.GetAllPastes(r.Context())
	if err != nil {
		log.Printf("get all pastes failed: %v", err)
		utils.RespondWithError(w, 500, "Error getting all pastes")
		return
	}
	pasteDataList := []map[string]string{}
	for _, paste := range pasteData {
		pasteDataList = append(pasteDataList, map[string]string{
			"id":         strconv.FormatInt(paste.ID, 10),
			"content":    paste.Content,
			"created_at": paste.CreatedAt.Time.Format(time.RFC3339),
			"updated_at": paste.UpdatedAt.Time.Format(time.RFC3339),
		})
	}

	utils.RespondWithJSON(w, http.StatusOK, pasteDataList)
}

func main() {
	_ = godotenv.Load("server/.env")
	_ = godotenv.Load(".env")

	if utils.IsLocalDev() && strings.TrimSpace(os.Getenv("DOMAIN")) == "" {
		log.Printf("local dev: short links use http://localhost:8081/{code} (set DOMAIN to override)")
	}

	servMux := http.NewServeMux()
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}).Handler(servMux)

	server := http.Server{
		Addr:    ":8081",
		Handler: corsHandler,
	}

	apiConfig := apiConfig{
		db:                    &database.Queries{},
		localDevDefaultDomain: "http://localhost:8081",
	}

	dbURL, err := utils.ResolveDBURL("/pastebin", "pastebin")
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

	servMux.HandleFunc("POST /api/pastebin/pastes", apiConfig.createPastebin)

	servMux.HandleFunc("GET /api/pastebin/pastes/{id}", apiConfig.getPastebin)

	servMux.HandleFunc("GET /api/pastebin/pastes", apiConfig.getAllPastes)

	servMux.HandleFunc("GET /api/pastebin/healthz", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	server.ListenAndServe()
}
