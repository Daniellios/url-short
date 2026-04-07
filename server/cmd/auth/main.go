package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/url-short/internal/auth/database"
	"example.com/url-short/internal/utils"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/oauth2"

	googleoauth2 "golang.org/x/oauth2/google"
)

const (
	oauthProviderGoogle = "google"

	cookieOAuthState = "oauth_state"
	cookieSession    = "auth_session"

	oauthStateMaxAgeSec = 600
)

type apiConfig struct {
	db *database.Queries

	oidcVerifier *oidc.IDTokenVerifier
	oauth2       *oauth2.Config

	sessionMaxAge time.Duration

	postLoginRedirectURL string
}

type googleClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func main() {
	_ = godotenv.Load("server/.env")
	_ = godotenv.Load(".env")

	clientID := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"))
	if clientID == "" || clientSecret == "" {
		log.Fatal("GOOGLE_OAUTH_CLIENT_ID and GOOGLE_OAUTH_CLIENT_SECRET are required")
	}

	publicBase := strings.TrimRight(strings.TrimSpace(os.Getenv("AUTH_PUBLIC_BASE_URL")), "/")
	if publicBase == "" {
		publicBase = "http://localhost:8082"
		log.Printf("AUTH_PUBLIC_BASE_URL not set; using %s", publicBase)
	}

	redirectURL := strings.TrimSpace(os.Getenv("AUTH_GOOGLE_REDIRECT_URL"))
	if redirectURL == "" {
		redirectURL = publicBase + "/api/auth/google/callback"
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		log.Fatalf("oidc provider: %v", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	oauthConf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     googleoauth2.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	sessionDays := 30
	if v := strings.TrimSpace(os.Getenv("AUTH_SESSION_DAYS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			sessionDays = n
		}
	}
	sessionMaxAge := time.Duration(sessionDays) * 24 * time.Hour

	postLogin := strings.TrimSpace(os.Getenv("AUTH_POST_LOGIN_REDIRECT"))
	if postLogin == "" {
		postLogin = "http://localhost:3000/"
		log.Printf("AUTH_POST_LOGIN_REDIRECT not set; using %s", postLogin)
	}

	addr := strings.TrimSpace(os.Getenv("AUTH_HTTP_ADDR"))
	if addr == "" {
		addr = ":8082"
	}

	servMux := http.NewServeMux()
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}).Handler(servMux)

	server := http.Server{
		Addr:    addr,
		Handler: corsHandler,
	}

	dbURL, err := utils.ResolveDBURL("/auth", "auth")
	if err != nil {
		log.Fatal(err)
	}
	if v := strings.TrimSpace(os.Getenv("AUTH_DB_URL")); v != "" {
		dbURL = v
	}

	dbQueries, dbConn, err := utils.SetupDB(dbURL, func(db *sql.DB) *database.Queries {
		return database.New(db)
	})
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer dbConn.Close()

	cfg := &apiConfig{
		db:                   dbQueries,
		oauth2:               oauthConf,
		oidcVerifier:         verifier,
		sessionMaxAge:        sessionMaxAge,
		postLoginRedirectURL: postLogin,
	}

	servMux.HandleFunc("GET /api/auth/google/start", cfg.handleGoogleStart)
	servMux.HandleFunc("GET /api/auth/google/callback", cfg.handleGoogleCallback)
	servMux.HandleFunc("GET /api/auth/me", cfg.handleMe)
	servMux.HandleFunc("POST /api/auth/logout", cfg.handleLogout)
	servMux.HandleFunc("GET /api/auth/healthz", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	log.Printf("auth listening on %s (Google callback %s)", addr, redirectURL)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func (cfg *apiConfig) handleGoogleStart(w http.ResponseWriter, r *http.Request) {
	state, err := randomURLToken(32)
	if err != nil {
		log.Printf("oauth state: %v", err)
		utils.RespondWithError(w, 500, "Could not start login")
		return
	}
	setCookie(w, cookieOAuthState, state, oauthStateMaxAgeSec)
	url := cfg.oauth2.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func (cfg *apiConfig) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" || stateCookie(r) != state {
		utils.RespondWithError(w, 400, "Invalid OAuth state")
		return
	}
	clearCookie(w, cookieOAuthState)

	code := r.URL.Query().Get("code")
	if code == "" {
		utils.RespondWithError(w, 400, "Missing authorization code")
		return
	}

	ctx := r.Context()
	token, err := cfg.oauth2.Exchange(ctx, code)
	if err != nil {
		log.Printf("oauth exchange: %v", err)
		utils.RespondWithError(w, 400, "OAuth exchange failed")
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		utils.RespondWithError(w, 400, "Missing id_token")
		return
	}

	idToken, err := cfg.oidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Printf("id_token verify: %v", err)
		utils.RespondWithError(w, 400, "Invalid id_token")
		return
	}

	var claims googleClaims
	if err := idToken.Claims(&claims); err != nil {
		log.Printf("claims: %v", err)
		utils.RespondWithError(w, 400, "Invalid claims")
		return
	}
	if claims.Sub == "" {
		utils.RespondWithError(w, 400, "Missing subject")
		return
	}

	userID, err := cfg.ensureGoogleUser(ctx, claims)
	if err != nil {
		log.Printf("ensure user: %v", err)
		utils.RespondWithError(w, 500, "Could not create session")
		return
	}

	sessionToken, err := randomURLToken(48)
	if err != nil {
		log.Printf("session token: %v", err)
		utils.RespondWithError(w, 500, "Could not create session")
		return
	}
	sum := sha256.Sum256([]byte(sessionToken))
	expires := time.Now().UTC().Add(cfg.sessionMaxAge)

	if _, err := cfg.db.CreateSession(ctx, database.CreateSessionParams{
		UserID:    userID,
		TokenHash: sum[:],
		ExpiresAt: expires,
	}); err != nil {
		log.Printf("create session: %v", err)
		utils.RespondWithError(w, 500, "Could not create session")
		return
	}

	maxAge := int(cfg.sessionMaxAge.Seconds())
	setCookie(w, cookieSession, sessionToken, maxAge)
	http.Redirect(w, r, cfg.postLoginRedirectURL, http.StatusFound)
}

func (cfg *apiConfig) ensureGoogleUser(ctx context.Context, claims googleClaims) (uuid.UUID, error) {
	ident, err := cfg.db.GetOAuthIdentity(ctx, database.GetOAuthIdentityParams{
		Provider:        oauthProviderGoogle,
		ProviderSubject: claims.Sub,
	})
	if err == nil {
		userID := ident.UserID
		if err := cfg.db.UpdateUserEmailAndDisplay(ctx, database.UpdateUserEmailAndDisplayParams{
			ID:          userID,
			Email:       nullString(claims.Email),
			DisplayName: nullString(claims.Name),
		}); err != nil {
			log.Printf("update user profile: %v", err)
		}
		return userID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, err
	}

	u, err := cfg.db.CreateUser(ctx, database.CreateUserParams{
		Email:       nullString(claims.Email),
		DisplayName: nullString(claims.Name),
	})
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return cfg.refetchGoogleUser(ctx, claims.Sub)
		}
		return uuid.Nil, err
	}
	if err := cfg.db.CreateOAuthIdentity(ctx, database.CreateOAuthIdentityParams{
		UserID:          u.ID,
		Provider:        oauthProviderGoogle,
		ProviderSubject: claims.Sub,
	}); err != nil {
		if utils.IsUniqueViolation(err) {
			return cfg.refetchGoogleUser(ctx, claims.Sub)
		}
		return uuid.Nil, err
	}
	return u.ID, nil
}

func (cfg *apiConfig) refetchGoogleUser(ctx context.Context, subject string) (uuid.UUID, error) {
	ident, err := cfg.db.GetOAuthIdentity(ctx, database.GetOAuthIdentityParams{
		Provider:        oauthProviderGoogle,
		ProviderSubject: subject,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return ident.UserID, nil
}

func (cfg *apiConfig) handleMe(w http.ResponseWriter, r *http.Request) {
	token := sessionCookieValue(r)
	if token == "" {
		utils.RespondWithError(w, 401, "Not authenticated")
		return
	}
	sum := sha256.Sum256([]byte(token))
	sess, err := cfg.db.GetSessionByTokenHash(r.Context(), sum[:])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, 401, "Not authenticated")
			return
		}
		log.Printf("session: %v", err)
		utils.RespondWithError(w, 500, "Session lookup failed")
		return
	}

	u, err := cfg.db.GetUser(r.Context(), sess.UserID)
	if err != nil {
		log.Printf("user: %v", err)
		utils.RespondWithError(w, 500, "User lookup failed")
		return
	}

	type userOut struct {
		ID          uuid.UUID `json:"id"`
		Email       *string   `json:"email,omitempty"`
		DisplayName *string   `json:"display_name,omitempty"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	out := userOut{ID: u.ID, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
	if u.Email.Valid {
		out.Email = &u.Email.String
	}
	if u.DisplayName.Valid {
		out.DisplayName = &u.DisplayName.String
	}
	utils.RespondWithJSON(w, http.StatusOK, out)
}

func (cfg *apiConfig) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := sessionCookieValue(r)
	if token != "" {
		sum := sha256.Sum256([]byte(token))
		if sess, err := cfg.db.GetSessionByTokenHash(r.Context(), sum[:]); err == nil {
			_ = cfg.db.DeleteSession(r.Context(), sess.ID)
		}
	}
	clearCookie(w, cookieSession)
	w.WriteHeader(http.StatusNoContent)
}

func nullString(s string) sql.NullString {
	if strings.TrimSpace(s) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func randomURLToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCookie(w http.ResponseWriter, name, value string, maxAgeSec int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAgeSec,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   utils.IsProduction(),
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   utils.IsProduction(),
	})
}

func stateCookie(r *http.Request) string {
	c, err := r.Cookie(cookieOAuthState)
	if err != nil {
		return ""
	}
	return c.Value
}

func sessionCookieValue(r *http.Request) string {
	c, err := r.Cookie(cookieSession)
	if err != nil {
		return ""
	}
	return c.Value
}
