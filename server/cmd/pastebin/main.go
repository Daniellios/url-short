package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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

type pasteResponse struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	IsPublic  bool    `json:"is_public"`
	ExpiresAt *string `json:"expires_at"`
}

func newPasteResponse(id int64, title, content string, createdAt, updatedAt, expiresAt sql.NullTime, isPublic bool) pasteResponse {
	out := pasteResponse{
		ID:       strconv.FormatInt(id, 10),
		Title:    title,
		Content:  content,
		IsPublic: isPublic,
	}
	if createdAt.Valid {
		out.CreatedAt = createdAt.Time.Format(time.RFC3339)
	}
	if updatedAt.Valid {
		out.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
	}
	if expiresAt.Valid {
		s := expiresAt.Time.Format(time.RFC3339)
		out.ExpiresAt = &s
	}
	return out
}

func (apiConfig *apiConfig) createPastebin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title     string  `json:"title"`
		Content   string  `json:"content"`
		IsPublic  *bool   `json:"is_public"`
		ExpiresAt *string `json:"expires_at"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, 400, "Error decoding parameters")
		return
	}

	title := strings.TrimSpace(params.Title)
	content := params.Content
	if strings.TrimSpace(content) == "" {
		utils.RespondWithError(w, 400, "Content is required")
		return
	}

	isPublic := true
	if params.IsPublic != nil {
		isPublic = *params.IsPublic
	}

	var expiresAtNT sql.NullTime
	if params.ExpiresAt != nil {
		raw := strings.TrimSpace(*params.ExpiresAt)
		if raw != "" {
			t, perr := time.Parse(time.RFC3339, raw)
			if perr != nil {
				t, perr = time.Parse(time.RFC3339Nano, raw)
			}
			if perr != nil {
				utils.RespondWithError(w, 400, "expires_at must be RFC3339 / ISO-8601 (e.g. 2026-12-31T15:04:05Z)")
				return
			}
			expiresAtNT = sql.NullTime{Time: t, Valid: true}
		}
	}

	log.Printf("title: %q content length: %d isPublic: %v expiresAt: %v", title, len(content), isPublic, expiresAtNT)

	createdPasteData, err := apiConfig.db.CreatePaste(r.Context(), database.CreatePasteParams{
		Title:     title,
		Content:   content,
		IsPublic:  isPublic,
		ExpiresAt: expiresAtNT,
	})
	if err != nil {
		log.Printf("create paste failed: %v", err)
		utils.RespondWithError(w, 500, "Error creating paste")
		return
	}
	log.Printf("created paste data: %+v", createdPasteData)
	utils.RespondWithJSON(w, http.StatusCreated, newPasteResponse(
		createdPasteData.ID,
		createdPasteData.Title,
		content,
		createdPasteData.CreatedAt,
		createdPasteData.UpdatedAt,
		createdPasteData.ExpiresAt,
		createdPasteData.IsPublic,
	))
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
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, 404, "Paste not found")
			return
		}
		log.Printf("get paste failed: %v", err)
		utils.RespondWithError(w, 500, "Error getting paste")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, newPasteResponse(
		pasteData.ID,
		pasteData.Title,
		pasteData.Content,
		pasteData.CreatedAt,
		pasteData.UpdatedAt,
		pasteData.ExpiresAt,
		pasteData.IsPublic,
	))

}

func (apiConfig *apiConfig) getAllPastes(w http.ResponseWriter, r *http.Request) {
	pasteData, err := apiConfig.db.GetAllPastes(r.Context())
	if err != nil {
		log.Printf("get all pastes failed: %v", err)
		utils.RespondWithError(w, 500, "Error getting all pastes")
		return
	}
	pasteDataList := make([]pasteResponse, 0, len(pasteData))
	for _, paste := range pasteData {
		pasteDataList = append(pasteDataList, newPasteResponse(
			paste.ID,
			paste.Title,
			paste.Content,
			paste.CreatedAt,
			paste.UpdatedAt,
			paste.ExpiresAt,
			paste.IsPublic,
		))
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
