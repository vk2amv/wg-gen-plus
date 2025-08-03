package model

import "time"

// User structure
type User struct {
	Sub      string    `json:"sub"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password,omitempty"` // omitempty prevents sending password in JSON responses
	IsAdmin  bool      `json:"isAdmin"`
	Profile  string    `json:"profile,omitempty"` // Keep but mark as omitempty
	Issuer   string    `json:"-"`                 // Hide completely from JSON
	IssuedAt time.Time `json:"-"`                 // Hide completely from JSON
}
