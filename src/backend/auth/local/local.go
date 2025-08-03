package local

import (
	"errors"
	"strings"
	"wg-gen-plus/model"
	"wg-gen-plus/storage"

	"golang.org/x/crypto/bcrypt"
)

// Local auth provider for username/password authentication
type Local struct{}

// Setup validate provider
func (l *Local) Setup() error {
	return nil // No setup needed now that we're using the database
}

// Authenticate checks username and password against database
func (l *Local) Authenticate(username, password string) (*model.User, error) {
	// Get all users from database
	users, err := storage.LoadAllUsers()
	if err != nil {
		return nil, errors.New("failed to load users from database")
	}

	// Find user by name (case insensitive)
	var foundUser *model.User
	for _, user := range users {
		// Use strings.EqualFold for case-insensitive comparison
		if strings.EqualFold(user.Name, username) {
			foundUser = user
			break
		}
	}

	// User not found
	if foundUser == nil {
		return nil, errors.New("invalid username or password")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Don't return the password hash
	foundUser.Password = ""

	return foundUser, nil
}
