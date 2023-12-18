package main

import (
	"context"
	"log"
	"mailer/api"
	googleauth "mailer/auth/google"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

func main() {
	apiServer := api.New()
	apiServer.Handle()
	apiServer.Listen()
	// program()
}

const (
	AZ = 10
)

func program() {
	ctx := context.Background()
	config := googleauth.GetConfig()
	token, err := googleauth.GetToken(config)
	if err != nil {
		log.Fatalln("Unable to get token: ", err)
	}
	client := config.Client(ctx, token)
	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalln("Unable to create Gmail service: ", err)
	}

	// Call the Gmail API to retrieve the most recent email
	user := "me"
	messages, err := gmailService.Users.Messages.List(user).MaxResults(10).Do()
	if err != nil {
		log.Fatalln("Unable to retrieve messages: ", err)
	}

	if len(messages.Messages) > 0 {
		for _, m := range messages.Messages {
			message, err := gmailService.Users.Messages.Get(user, m.Id).Do()
			if err != nil {
				log.Fatalln("Unable to retrieve message details: ", err)
			}
			log.Println("Message snippet:", message.Snippet)
		}
	} else {
		log.Println("No messages found.")
	}

	peopleService, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalln("Unable to create People service: ", err)
	}
	person, err := peopleService.People.Get("people/me").PersonFields("emailAddresses,names").Do()
	if err != nil {
		log.Fatalln("Unable to get user info: ", err)
	}

	log.Println("User email", person.EmailAddresses[0].Value)
	log.Println("User email", person.Names[0].DisplayName)
}
