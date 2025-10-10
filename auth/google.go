package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var Scopes = []string{
	"https://www.googleapis.com/auth/drive.file",
	"https://www.googleapis.com/auth/calendar",
}

// GetServiceClient — полностью безбраузерный OAuth-клиент.
// Требует заранее созданные файлы secrets/client_secret.json и secrets/google-token.json.
func GetServiceClient() *http.Client {
	ctx := context.Background()

	// 1️⃣ Пути к секретам
	clientSecretPath := filepath.Join("secrets", "client_secret.json")
	tokenPath := filepath.Join("secrets", "google-token.json")

	// 2️⃣ Загружаем client_secret.json
	b, err := os.ReadFile(clientSecretPath)
	if err != nil {
		log.Fatalf("Unable to read %s: %v", clientSecretPath, err)
	}

	config, err := google.ConfigFromJSON(b, Scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret: %v", err)
	}

	// 3️⃣ Загружаем токен (создан вручную через manual_auth.go)
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		log.Fatalf("Missing or invalid token file (%s): %v", tokenPath, err)
	}

	// 4️⃣ Проверяем срок действия токена
	ts := config.TokenSource(ctx, tok)
	newTok, err := ts.Token()
	if err != nil {
		log.Fatalf("Unable to refresh token: %v", err)
	}

	// 5️⃣ Если access_token обновился — сохраняем новый
	if newTok.AccessToken != tok.AccessToken {
		saveToken(tokenPath, newTok)
	}

	return config.Client(ctx, newTok)
}

// ------------------ вспомогательные функции ------------------

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	f, err := os.Create(path)
	if err != nil {
		log.Printf("⚠️ Unable to save refreshed token: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	fmt.Printf("💾 Token updated in %s\n", path)
}
