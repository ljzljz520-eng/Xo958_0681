package shop

import "errors"

var (
	ErrEmailInUse      = errors.New("email already in use")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidLogin    = errors.New("invalid login")
	ErrNotFound        = errors.New("not found")
	ErrUnauthenticated = errors.New("unauthenticated")
)

type Soap struct {
	Slug       string
	Name       string
	Scent      string
	Story      string
	PriceCents int
}

type Account struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
}

type NewAccount struct {
	Name         string
	Email        string
	PasswordHash string
}

type MemberNote interface {
	Title() string
	Body() string
}

type MemberPage struct {
	Name      string
	Empty     bool
	NoteTitle string
	NoteBody  string
}

type CatalogRepository interface {
	List() []Soap
	FindBySlug(slug string) (Soap, bool)
}

type AccountRepository interface {
	Create(input NewAccount) (Account, error)
	FindByEmail(email string) (Account, bool)
	FindByID(id string) (Account, bool)
}

type MemberNoteRepository interface {
	FindByAccountID(accountID string) (MemberNote, error)
}

type SessionRepository interface {
	Create(accountID string) string
	AccountID(token string) (string, bool)
	Delete(token string)
}
