package auth

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"time"
	"wg-gen-plus/auth"
	"wg-gen-plus/model"
	"wg-gen-plus/util"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/auth")
	{
		g.GET("/type", getAuthType) // Add this line
		g.GET("/oauth2_url", oauth2URL)
		g.POST("/oauth2_exchange", oauth2Exchange)
		g.POST("/login", handleLocalLogin) // Add this if not already present
		g.GET("/logout", logout)
		g.GET("/user", user)
	}
}

/*
 * generate redirect url to get OAuth2 code or let client know that OAuth2 is disabled
 */
func oauth2URL(c *gin.Context) {
	// First check if local auth is enabled
	if auth.IsLocalAuth() {
		c.JSON(http.StatusOK, &model.Auth{
			Oauth2:   false,
			ClientId: "",
			State:    "",
			CodeUrl:  "",
		})
		return
	}

	cacheDb := c.MustGet("cache").(*cache.Cache)

	state, err := util.GenerateRandomString(32)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to generate state random string")
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	clientId, err := util.GenerateRandomString(32)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to generate state random string")
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	// save clientId and state so we can retrieve for verification
	cacheDb.Set(clientId, state, 5*time.Minute)

	oauth2Client := c.MustGet("oauth2Client").(auth.Auth)

	data := &model.Auth{
		Oauth2:   true,
		ClientId: clientId,
		State:    state,
		CodeUrl:  oauth2Client.CodeUrl(state),
	}

	c.JSON(http.StatusOK, data)
}

/*
 * exchange code and get user infos, if OAuth2 is disable just send fake data
 */
func oauth2Exchange(c *gin.Context) {
	// First check if local auth is enabled
	if auth.IsLocalAuth() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth2 is disabled when using local authentication"})
		return
	}

	var loginVals model.Auth
	if err := c.ShouldBind(&loginVals); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("code and state fields are missing")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	cacheDb := c.MustGet("cache").(*cache.Cache)
	savedState, exists := cacheDb.Get(loginVals.ClientId)

	if !exists || savedState != loginVals.State {
		log.WithFields(log.Fields{
			"state":      loginVals.State,
			"savedState": savedState,
		}).Error("saved state and client provided state mismatch")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	oauth2Client := c.MustGet("oauth2Client").(auth.Auth)

	oauth2Token, err := oauth2Client.Exchange(loginVals.Code)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to exchange code for token")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	cacheDb.Delete(loginVals.ClientId)
	cacheDb.Set(oauth2Token.AccessToken, oauth2Token, cache.DefaultExpiration)

	c.JSON(http.StatusOK, oauth2Token.AccessToken)
}

func logout(c *gin.Context) {
	cacheDb := c.MustGet("cache").(*cache.Cache)
	cacheDb.Delete(c.Request.Header.Get(util.AuthTokenHeaderName))
	c.JSON(http.StatusOK, gin.H{})
}

func user(c *gin.Context) {
	cacheDb := c.MustGet("cache").(*cache.Cache)
	oauth2Token, exists := cacheDb.Get(c.Request.Header.Get(util.AuthTokenHeaderName))

	if exists && oauth2Token.(*oauth2.Token).AccessToken == c.Request.Header.Get(util.AuthTokenHeaderName) {
		oauth2Client := c.MustGet("oauth2Client").(auth.Auth)
		user, err := oauth2Client.UserInfo(oauth2Token.(*oauth2.Token))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to get user from oauth2 AccessToken")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.JSON(http.StatusOK, user)
		return
	}

	log.WithFields(log.Fields{
		"exists":                 exists,
		util.AuthTokenHeaderName: c.Request.Header.Get(util.AuthTokenHeaderName),
	}).Error("oauth2 AccessToken is not recognized")

	c.AbortWithStatus(http.StatusUnauthorized)
}

// Local login handler
func handleLocalLogin(c *gin.Context) {
	var loginData struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}

	// Log raw request for debugging
	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Redact password from logs - only log that a request was received
	log.Infof("Login request received, Content-Type: %s", c.GetHeader("Content-Type"))

	// Try to bind JSON first, then form data
	if err := c.ShouldBindJSON(&loginData); err != nil {
		log.Warnf("JSON binding error: %v", err)
		if err := c.ShouldBind(&loginData); err != nil {
			log.Warnf("Form binding error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password required"})
			return
		}
	}

	log.Infof("Login attempt for user: %s (password redacted)", loginData.Username)

	username := loginData.Username
	password := loginData.Password

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password required"})
		return
	}

	localAuth, err := auth.GetLocalAuthProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth provider not available"})
		return
	}

	user, err := localAuth.Authenticate(username, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Generate token
	token, err := util.GenerateRandomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate auth token"})
		return
	}

	// Store token in cache
	cacheDb := c.MustGet("cache").(*cache.Cache)
	cacheDb.Set(token, user.Sub, 24*time.Hour) // Token valid for 24 hours

	// Return token and user info
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"sub":     user.Sub,
			"name":    user.Name,
			"email":   user.Email,
			"isAdmin": user.IsAdmin,
		},
	})
}

// Add this function at the end of the file
func getAuthType(c *gin.Context) {
	isLocal := auth.IsLocalAuth()

	c.JSON(http.StatusOK, gin.H{
		"isLocal":  isLocal,
		"authType": os.Getenv("AUTH_TYPE"),
	})
}
