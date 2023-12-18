package receiver

import (
	"fmt"
	"log"
	"mailer/services"

	"github.com/emersion/go-imap/client"
)

type LoginAuth struct {
	Username    string
	Password    string
	MailService services.MailService
}

func NewLoginAuth(username string, password string, mailService services.MailService) LoginAuth {
	return LoginAuth{
		Username:    username,
		Password:    password,
		MailService: mailService,
	}
}

func (auth *LoginAuth) Login() (*client.Client, error) {
	c, err := client.DialTLS(fmt.Sprintf("%s:%d", auth.MailService.ImapServer.Host, auth.MailService.ImapServer.TlsPort), nil)
	if err != nil {
		log.Fatalln("Unable to initialize imap client: ", err)
	}
	defer c.Logout()

	if err := c.Login(auth.Username, auth.Password); err != nil {
		log.Println("Got error: ", err)
		return nil, err
	}
	return nil, nil
}

func FetchEmails(c *client.Client) error {
	return nil
}
