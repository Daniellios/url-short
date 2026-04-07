package utils

import (
	"os"
	"strings"
)

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
