package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	googleauth "mailer/auth/google"
	"mailer/statestore"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/asdine/storm/v3/q"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

type ApiServer struct{}

var (
	ss           = statestore.New("./statestore.db")
	oauth2Config *oauth2.Config
	listener     net.Listener
	apiPort      int
)

func returnToken(w http.ResponseWriter, token *oauth2.Token) {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(tokenJSON)
}

// handle login for google account
func googleLogin(w http.ResponseWriter, r *http.Request) {
	oauth2Config = googleauth.GetConfig()
	oauth2Config.RedirectURL = fmt.Sprintf("http://localhost:%d/google/callback", apiPort)
	url := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Println("Auth url: ", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// handle oauth2 redirect after authentication
func googleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != "state-token" {
		log.Println("Invalid oauth2 state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Println("Code exchange failed: ", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	peopleService, err := people.NewService(r.Context(), option.WithHTTPClient(client))
	if err != nil {
		log.Println("Failed to create People service: ", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	person, err := peopleService.People.Get("people/me").PersonFields("emailAddresses,names").Do()
	if err != nil {
		log.Println("Failed to get user info: ", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	account := statestore.Account{
		Email:       person.EmailAddresses[0].Value,
		AccountType: "google",
		Token:       *token,
		UpdatedAt:   time.Now(),
		Scopes:      googleauth.DEFAULT_SCOPES,
	}
	account.DistinctID = fmt.Sprintf("%s+%s", account.AccountType, account.Email)

	ss.Open()
	defer ss.Close()
	if !ss.IsExistAccount(account) {
		fmt.Fprintf(w, "Added new account %s", account.Email)
	} else {
		fmt.Fprintf(w, "Account %s already existed", account.Email)
	}
	// upsert account to statestore
	ss.InternalDB().Save(&account)
	// json.NewEncoder(w).Encode(account)
}

// load token from statestore first, then refresh and update if it is expiried and return the latest token
func googleToken(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	accountType := r.FormValue("account_type")

	ss.Open()
	defer ss.Close()
	var account statestore.Account
	err := ss.InternalDB().One("DistinctID", fmt.Sprintf("%s+%s", accountType, email), &account)
	if err != nil {
		log.Println("Error when retrieving account from statestore: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if googleauth.IsTokenExpiried(&account.Token) {
		newToken, err := googleauth.RefreshToken(&account.Token)
		if err != nil {
			log.Println("Unable to refresh token: ", err)
		}
		if newToken.AccessToken != account.Token.AccessToken {
			log.Printf("Refresh token successfully for %s %s\n", accountType, email)
			account.Token = *newToken
			ss.InternalDB().Save(&account)
		}
	}
	json.NewEncoder(w).Encode(account.Token)
}

func fetchAccounts(w http.ResponseWriter, r *http.Request) {
	accountType := r.FormValue("account_type")
	email := r.FormValue("email")

	var accounts []statestore.Account
	ss.Open()
	defer ss.Close()

	// make a query
	var accountQuery *q.Matcher = nil
	emailQuery := q.Eq("Email", email)
	accountTypeQuery := q.Eq("AccountType", accountType)
	if email != "" && accountType != "" && accountType != "all" {
		aq := q.And(emailQuery, accountTypeQuery)
		accountQuery = &aq
	} else if email != "" {
		accountQuery = &emailQuery
	} else if accountType != "" && accountType != "all" {
		accountQuery = &accountTypeQuery
	}

	var err error
	if accountQuery != nil {
		err = ss.InternalDB().Select(*accountQuery).Find(&accounts)
	} else {
		err = ss.InternalDB().All(&accounts)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := struct {
		Data []statestore.Account `json:"data"`
	}{Data: accounts}
	json.NewEncoder(w).Encode(data)
}

func retrieveGmailMessages(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.FormValue("n"))
	if err != nil {
		log.Printf("Unable to parse n to int: %v, use default num = 10", err)
		n = 10
	}

	resp, err := http.Get(fmt.Sprintf("http://:%d/google/token?%s", apiPort, r.Form.Encode()))
	if err != nil {
		log.Fatalln("Error when making request: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalln("Unexpected status code: ", resp.StatusCode)
	}

	token := &oauth2.Token{}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		log.Fatalln("Error decoding token response: ", err)
	}

	gmailService, err := gmail.NewService(context.TODO(), option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		log.Fatalln("Unable to create Gmail service: ", err)
	}

	// Call the Gmail API to retrieve the most recent email
	listMessages, err := gmailService.Users.Messages.List("me").MaxResults(int64(n)).Do()
	if err != nil {
		log.Fatalln("Unable to retrieve messages: ", err)
	}
	messages, err := handleGmailMessages(gmailService, listMessages)
	if err != nil {
		log.Println("Got something wrong when fetching messages: ", err)
	}
	data := struct {
		Data []*gmail.Message `json:"data"`
	}{Data: messages}
	json.NewEncoder(w).Encode(&data)
}

func handleGmailMessages(gmailService *gmail.Service, messages *gmail.ListMessagesResponse) ([]*gmail.Message, error) {
	result := []*gmail.Message{}
	var err error = nil
	if len(messages.Messages) > 0 {
		processMessage := func(m *gmail.Message, wg *sync.WaitGroup) {
			defer wg.Done()
			message, err := gmailService.Users.Messages.Get("me", m.Id).Do()
			if err != nil {
				log.Fatalln("Unable to retrieve message details: ", err)
			}
			result = append(result, message)
		}
		var wg sync.WaitGroup
		for _, m := range messages.Messages {
			wg.Add(1)
			go processMessage(m, &wg)
		}
		wg.Wait()
	}
	return result, err
}

func New() *ApiServer {
	return &ApiServer{}
}

func (s *ApiServer) Handle() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "<h1>Mailer says hello</h1>")
	})

	http.HandleFunc("/google/login", googleLogin)
	http.HandleFunc("/google/callback", googleCallback)
	http.HandleFunc("/google/token", googleToken)
	http.HandleFunc("/google/gmail/messages", retrieveGmailMessages)
	http.HandleFunc("/accounts", fetchAccounts)
}

func createListener() net.Listener {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	return l
}

func (s *ApiServer) Listen() {
	listener = createListener()
	apiPort = listener.Addr().(*net.TCPAddr).Port
	defer listener.Close()
	log.Println("listening at", listener.Addr().String())
	http.Serve(listener, nil)
}
