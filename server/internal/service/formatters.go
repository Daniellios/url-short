package service

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

const localDevDefaultDomain = "http://localhost:8080"

func IsProduction() bool {
	if strings.TrimSpace(os.Getenv("DOCKER")) == "1" {
		return true
	}
	for _, key := range []string{"APP_ENV", "ENV"} {
		if strings.ToLower(strings.TrimSpace(os.Getenv(key))) == "production" {
			return true
		}
	}
	return false
}

func IsLocalDev() bool {
	if IsProduction() {
		return false
	}
	for _, key := range []string{"APP_ENV", "ENV"} {
		switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
		case "development", "dev", "local":
			return true
		}
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("DEV"))) {
	case "1", "true", "yes":
		return true
	}
	return true
}

func FormatShortURL(shortURL string) string {
	domain := strings.TrimRight(strings.TrimSpace(os.Getenv("DOMAIN")), "/")
	if domain != "" {
		return domain + "/" + shortURL
	}
	if IsLocalDev() {
		return localDevDefaultDomain + "/" + shortURL
	}
	return "/" + shortURL
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
