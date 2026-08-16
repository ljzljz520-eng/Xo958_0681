package shop

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

type Service struct {
	catalog  CatalogRepository
	accounts AccountRepository
	notes    MemberNoteRepository
	sessions SessionRepository
}

func NewService(catalog CatalogRepository, accounts AccountRepository, notes MemberNoteRepository, sessions SessionRepository) *Service {
	return &Service{catalog: catalog, accounts: accounts, notes: notes, sessions: sessions}
}

func HashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func (s *Service) Soaps() []Soap {
	return s.catalog.List()
}

func (s *Service) Soap(slug string) (Soap, error) {
	soap, ok := s.catalog.FindBySlug(slug)
	if !ok {
		return Soap{}, ErrNotFound
	}
	return soap, nil
}

func (s *Service) Register(name, email, password string) error {
	name = strings.TrimSpace(name)
	email = normalizeEmail(email)
	if name == "" || !strings.Contains(email, "@") || len(password) < 8 {
		return ErrInvalidInput
	}
	_, err := s.accounts.Create(NewAccount{Name: name, Email: email, PasswordHash: HashPassword(password)})
	return err
}

func (s *Service) Login(email, password string) (string, error) {
	account, ok := s.accounts.FindByEmail(normalizeEmail(email))
	if !ok || subtle.ConstantTimeCompare([]byte(account.PasswordHash), []byte(HashPassword(password))) != 1 {
		return "", ErrInvalidLogin
	}
	return s.sessions.Create(account.ID), nil
}

func (s *Service) Logout(token string) {
	s.sessions.Delete(token)
}

func (s *Service) Viewer(token string) (Account, bool) {
	accountID, ok := s.sessions.AccountID(token)
	if !ok {
		return Account{}, false
	}
	return s.accounts.FindByID(accountID)
}

func (s *Service) Member(token string) (MemberPage, error) {
	account, ok := s.Viewer(token)
	if !ok {
		return MemberPage{}, ErrUnauthenticated
	}
	note, err := s.notes.FindByAccountID(account.ID)
	if err != nil {
		return MemberPage{}, err
	}
	if note == nil {
		return MemberPage{Name: account.Name, Empty: true}, nil
	}
	return MemberPage{Name: account.Name, NoteTitle: note.Title(), NoteBody: note.Body()}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
