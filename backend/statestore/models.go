package statestore

import (
	"time"

	"golang.org/x/oauth2"
)

type Account struct {
	DistinctID  string       `json:"distinct_id" storm:"id"`
	Email       string       `json:"email" storm:"index"`
	AccountType string       `json:"account_type" storm:"index"`
	Token       oauth2.Token `json:"token" storm:"inline"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Scopes      []string     `json:"scopes"`
}
