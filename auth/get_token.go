package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var scopes = []string{
	"https://www.googleapis.com/auth/drive.file",
	"https://www.googleapis.com/auth/calendar",
}

func Manual_auth() {
	// --- Пути к файлам ---
	secretsDir := "secrets"
	clientSecretPath := filepath.Join(secretsDir, "client_secret.json")
	tokenPath := filepath.Join(secretsDir, "google-token.json")

	// --- Читаем client_secret.json ---
	b, err := os.ReadFile(clientSecretPath)
	if err != nil {
		log.Fatalf("Unable to read %s: %v", clientSecretPath, err)
	}

	// --- Создаём OAuth2 конфигурацию ---
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret: %v", err)
	}

	// --- Генерируем ссылку для авторизации ---
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("\n🔗 Open the following link in your browser and authorize the app:\n%v\n", authURL)

	fmt.Print("\nEnter the authorization code: ")
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	// --- Получаем токен ---
	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token: %v", err)
	}

	// --- Сохраняем токен в secrets/google-token.json ---
	os.MkdirAll(secretsDir, 0700)
	f, err := os.Create(tokenPath)
	if err != nil {
		log.Fatalf("Unable to create token file: %v", err)
	}
	defer f.Close()

	json.NewEncoder(f).Encode(tok)
	fmt.Printf("✅ Token saved to %s\n", tokenPath)
}
