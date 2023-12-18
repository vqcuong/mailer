package googleauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/mpvl/unique"
	"github.com/toqueteos/webbrowser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/people/v1"
)

const (
	tokenFile       = "./secrets/google/token.json"
	credentialsFile = "./secrets/google/credentials.json"
)

var DEFAULT_SCOPES = []string{
	gmail.GmailReadonlyScope,
	people.UserinfoEmailScope,
	people.UserinfoProfileScope,
	// people.UserEmailsReadScope,
}

func GetConfig(scope ...string) *oauth2.Config {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalln("Unable to get oauth2 config: ", err)
		return nil
	}
	finalScopes := append(DEFAULT_SCOPES, scope...)
	unique.Strings(&finalScopes)
	log.Println("Requesting scopes: ", finalScopes)
	config, err := google.ConfigFromJSON(b, finalScopes...)
	if err != nil {
		log.Fatalln("Unable to get oauth2 config: ", err)
	}
	return config
}

func RefreshToken(token *oauth2.Token) (*oauth2.Token, error) {
	config := GetConfig()
	tokenSource := config.TokenSource(context.TODO(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

func IsTokenExpiried(token *oauth2.Token) bool {
	// If the expiry time is zero, the token never expires
	if token.Expiry.IsZero() {
		return false
	}
	// Compare the expiry time with the current time
	return token.Expiry.Before(time.Now())
}

// ------------------------------- Only suitable for testing ----------------------------------

// Retrieve a token, saves the token, then returns the generated client.
func GetToken(config *oauth2.Config) (*oauth2.Token, error) {
	token, err := ReadToken()
	if err != nil {
		return GetTokenFromWeb(config)
	}
	if IsTokenExpiried(token) {
		return RefreshToken(token)
	}
	return token, err
}

// Request a token from the web, then returns the retrieved token.
func GetTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	config.RedirectURL = "http://localhost:3000/google/callback"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Println("Go to the following link in your browser then type the authorization code: ", authURL)
	err := webbrowser.Open(authURL)
	if err != nil {
		log.Printf("Unable open authURL automatically on webbrowser due to: %v, please do it manually\n", err)
	}
	HandleOauth2Redirect(config)
	return ReadToken()
}

func HandleOauth2Redirect(config *oauth2.Config) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := &http.Server{Addr: ":3000"}

	http.HandleFunc("/google/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		fmt.Fprintf(w, "Received authenticated code: %v", code)
		token, err := config.Exchange(context.TODO(), code)
		if err != nil {
			log.Fatalln("Unable to retrieve token from web: ", err)
		}
		SaveToken(token)
		// Shut down the server when this endpoint is called
		go func() {
			log.Println("Handled a redirected call, shutting down the server...")
			if err := server.Shutdown(context.TODO()); err != nil {
				log.Fatalln("Error shutting down server: ", err)
			}
		}()
	})

	// Start the server in a separate Goroutine
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("Server error: ", err)
		}
	}()

	wg.Wait()
}

// read token from file, then return an oauth2.Token pointer
func ReadToken() (*oauth2.Token, error) {
	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// Saves a token to a file path.
func SaveToken(token *oauth2.Token) {
	log.Println("Saving credential file to: ", tokenFile)
	f, err := os.Create(tokenFile)
	if err != nil {
		log.Fatalln("Unable to cache oauth token: ", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
