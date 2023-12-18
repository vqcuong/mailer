package statestore

import (
	"strings"

	"github.com/asdine/storm/v3"
)

type StateStore struct {
	db   *storm.DB
	path string
}

func New(path string) *StateStore {
	return &StateStore{path: path}
}

func (s *StateStore) Open(opts ...func(*storm.Options) error) error {
	db, err := storm.Open(s.path, opts...)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *StateStore) InternalDB() *storm.DB {
	return s.db
}

func (s *StateStore) Close() {
	s.db.Close()
}

func (s *StateStore) IsExistAccount(account Account) bool {
	var a Account
	err := s.db.One("DistinctID", strings.Join([]string{account.AccountType, account.Email}, "+"), &a)
	return err == nil
}
