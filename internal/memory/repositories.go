package memory

import (
	"fmt"
	"sync"

	"handmade-soap-shop/internal/shop"
)

type Catalog struct {
	soaps []shop.Soap
}

func NewCatalog(soaps []shop.Soap) *Catalog {
	return &Catalog{soaps: append([]shop.Soap(nil), soaps...)}
}

func (c *Catalog) List() []shop.Soap {
	return append([]shop.Soap(nil), c.soaps...)
}

func (c *Catalog) FindBySlug(slug string) (shop.Soap, bool) {
	for _, soap := range c.soaps {
		if soap.Slug == slug {
			return soap, true
		}
	}
	return shop.Soap{}, false
}

type Accounts struct {
	mu      sync.RWMutex
	byID    map[string]shop.Account
	byEmail map[string]string
	nextID  int
}

func NewAccounts(seed []shop.Account) *Accounts {
	r := &Accounts{byID: make(map[string]shop.Account), byEmail: make(map[string]string), nextID: len(seed) + 1}
	for _, account := range seed {
		r.byID[account.ID] = account
		r.byEmail[account.Email] = account.ID
	}
	return r
}

func (r *Accounts) Create(input shop.NewAccount) (shop.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byEmail[input.Email]; exists {
		return shop.Account{}, shop.ErrEmailInUse
	}
	var id string
	for {
		id = fmt.Sprintf("member-%03d", r.nextID)
		r.nextID++
		if _, exists := r.byID[id]; !exists {
			break
		}
	}
	account := shop.Account{ID: id, Name: input.Name, Email: input.Email, PasswordHash: input.PasswordHash}
	r.byID[id] = account
	r.byEmail[input.Email] = id
	return account, nil
}

func (r *Accounts) FindByEmail(email string) (shop.Account, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return shop.Account{}, false
	}
	account, ok := r.byID[id]
	return account, ok
}

func (r *Accounts) FindByID(id string) (shop.Account, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	account, ok := r.byID[id]
	return account, ok
}

type noteRow struct {
	title string
	body  string
}

func (r *noteRow) Title() string {
	return r.title
}

func (r *noteRow) Body() string {
	return r.body
}

type MemberNotes struct {
	mu      sync.RWMutex
	records map[string]*noteRow
}

type MemberNoteSeed struct {
	AccountID string
	Title     string
	Body      string
}

func NewMemberNotes(seed []MemberNoteSeed) *MemberNotes {
	records := make(map[string]*noteRow, len(seed))
	for _, item := range seed {
		records[item.AccountID] = &noteRow{title: item.Title, body: item.Body}
	}
	return &MemberNotes{records: records}
}

func (r *MemberNotes) FindByAccountID(accountID string) (shop.MemberNote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record := r.records[accountID]
	return record, nil
}

type Sessions struct {
	mu        sync.RWMutex
	accountBy map[string]string
	nextID    int
}

func NewSessions() *Sessions {
	return &Sessions{accountBy: make(map[string]string), nextID: 1}
}

func (r *Sessions) Create(accountID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	token := fmt.Sprintf("session-%06d", r.nextID)
	r.nextID++
	r.accountBy[token] = accountID
	return token
}

func (r *Sessions) AccountID(token string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	accountID, ok := r.accountBy[token]
	return accountID, ok
}

func (r *Sessions) Delete(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.accountBy, token)
}
