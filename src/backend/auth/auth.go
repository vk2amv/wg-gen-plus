package auth

import (
	"fmt"
	"os"
	"strings"
	"wg-gen-plus/auth/fake"
	"wg-gen-plus/auth/github"
	"wg-gen-plus/auth/local"
	"wg-gen-plus/auth/oauth2oidc"
	"wg-gen-plus/model"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Auth interface to implement as auth provider
type Auth interface {
	Setup() error
	CodeUrl(state string) string
	Exchange(code string) (*oauth2.Token, error)
	UserInfo(oauth2Token *oauth2.Token) (*model.User, error)
}

// LocalAuth interface for username/password authentication
type LocalAuth interface {
	Setup() error
	Authenticate(username, password string) (*model.User, error)
}

// IsLocalAuth returns true if auth type is local
func IsLocalAuth() bool {
	return strings.ToLower(os.Getenv("AUTH_TYPE")) == "local"
}

// GetAuthProvider get an instance of oauth2 auth provider
func GetAuthProvider() (Auth, error) {
	var oauth2Client Auth
	var err error

	switch os.Getenv("OAUTH2_PROVIDER_NAME") {
	case "fake":
		log.Warn("Oauth is set to fake, no actual authentication will be performed")
		oauth2Client = &fake.Fake{}

	case "oauth2oidc":
		log.Warn("Oauth is set to oauth2oidc, must be RFC implementation on server side")
		oauth2Client = &oauth2oidc.Oauth2idc{}

	case "github":
		log.Warn("Oauth is set to github, no openid will be used")
		oauth2Client = &github.Github{}

	default:
		return nil, fmt.Errorf("auth provider name %s unknown", os.Getenv("OAUTH2_PROVIDER_NAME"))
	}

	err = oauth2Client.Setup()
	return oauth2Client, err
}

// GetLocalAuthProvider get an instance of local auth provider
func GetLocalAuthProvider() (LocalAuth, error) {
	localClient := &local.Local{}
	err := localClient.Setup()
	return localClient, err
}
