package services

type Server struct {
	Host    string
	TlsPort int
}

type MailService struct {
	Name        string
	DisplayName string
	ImapServer  Server
	SmtpServer  Server
}

var (
	GMAIL     = MailService{Name: "gmail", DisplayName: "Gmail", ImapServer: Server{Host: "imap.gmail.com", TlsPort: 993}, SmtpServer: Server{Host: "smtp.gmail.com", TlsPort: 587}}
	OFFICE365 = MailService{Name: "office365", DisplayName: "Office365", ImapServer: Server{Host: "outlook.office365.com", TlsPort: 993}, SmtpServer: Server{Host: "smtp.office365.com", TlsPort: 587}}
	YAHOO     = MailService{Name: "yahoo", DisplayName: "yahoo", ImapServer: Server{Host: "imap.mail.yahoo.com", TlsPort: 587}, SmtpServer: Server{Host: "smtp.mail.yahoo.com", TlsPort: 587}}
)

var (
	GOOGLE_CLIENT_ID     = "193536113064-0jo157v5abq9hcc4ob3ac9ij8c6jsf12.apps.googleusercontent.com"
	GOOGLE_CLIENT_SECRET = "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"
)
