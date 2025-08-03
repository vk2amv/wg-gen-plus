package core

import (
	"errors"
	"strings"
	"wg-gen-plus/model"
	"wg-gen-plus/storage"

	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
)

// CreateUser creates a new user
func CreateUser(user *model.User) (*model.User, error) {
	// Generate a unique ID if not provided
	if user.Sub == "" {
		u, err := uuid.NewV4()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to generate UUID")
			return nil, errors.New("failed to generate user ID")
		}
		user.Sub = u.String()
	}

	// Basic validation
	if user.Name == "" {
		return nil, errors.New("user name cannot be empty")
	}

	// Check if a user with this name already exists
	existingUsers, err := storage.LoadAllUsers()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to check for existing users")
		return nil, err
	}

	for _, existingUser := range existingUsers {
		if strings.EqualFold(existingUser.Name, user.Name) && existingUser.Sub != user.Sub {
			return nil, errors.New("a user with this name already exists")
		}
	}

	// Save to database
	err = storage.SaveUser(user)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to save user")
		return nil, err
	}

	// Reload from DB to ensure all fields are set
	return storage.LoadUser(user.Sub)
}

// ReadUser retrieves a user by their ID
func ReadUser(id string) (*model.User, error) {
	return storage.LoadUser(id)
}

// ReadUsers retrieves all users
func ReadUsers() ([]*model.User, error) {
	return storage.LoadAllUsers()
}

// UpdateUser updates an existing user
func UpdateUser(id string, user *model.User) (*model.User, error) {
	// Make sure the user exists
	_, err := storage.LoadUser(id)
	if err != nil {
		return nil, err
	}

	// Prevent changing the ID
	if user.Sub != id {
		return nil, errors.New("cannot change user ID")
	}

	// Basic validation
	if user.Name == "" {
		return nil, errors.New("user name cannot be empty")
	}

	// Check if another user with this name already exists
	existingUsers, err := storage.LoadAllUsers()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to check for existing users")
		return nil, err
	}

	for _, existingUser := range existingUsers {
		if strings.EqualFold(existingUser.Name, user.Name) && existingUser.Sub != user.Sub {
			return nil, errors.New("another user with this name already exists")
		}
	}

	// Save changes
	err = storage.SaveUser(user)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update user")
		return nil, err
	}

	// Reload from DB to ensure all fields are set
	return storage.LoadUser(id)
}

// DeleteUser removes a user
func DeleteUser(id string) error {
	err := storage.DeleteUser(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to delete user")
		return err
	}
	return nil
}
