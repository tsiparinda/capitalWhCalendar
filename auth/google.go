package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Scopes — список API-доступов, которые мы разрешаем одновременно.
var Scopes = []string{
	"https://www.googleapis.com/auth/drive.file",   // доступ к Drive (только файлы, созданные приложением)
	"https://www.googleapis.com/auth/calendar",     // доступ к Google Calendar
}

// GetClient — возвращает общий HTTP client для всех сервисов (Drive, Calendar).
func GetClient() *http.Client {
	ctx := context.Background()

	b, err := os.ReadFile("../secrets/client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client_secret.json: %v", err)
	}

	config, err := google.ConfigFromJSON(b, Scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file: %v", err)
	}

	tokenFile := tokenFilePath()
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		fmt.Println("⚠️  No valid token found. Starting new authorization flow...")
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	} else {
		ts := config.TokenSource(ctx, tok)
		newTok, err := ts.Token()
		if err != nil {
			fmt.Println("⚠️  Token refresh failed:", err)
			fmt.Println("🔄  Starting new authorization flow...")
			tok = getTokenFromWeb(config)
			saveToken(tokenFile, tok)
		} else if newTok.AccessToken != tok.AccessToken {
			saveToken(tokenFile, newTok)
			tok = newTok
		}
	}

	return config.Client(ctx, tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Open the following link in your browser and authorize the application:\n%v\n", authURL)

	fmt.Print("Enter the authorization code: ")
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to get user home directory: %v", err)
	}
	tokenDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenDir, 0700)
	return filepath.Join(tokenDir, "google-token.json")
}

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
	fmt.Printf("💾 Saving token to %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
