package users

import (
	"net/http"
	"wg-gen-plus/auth"
	"wg-gen-plus/core"
	"wg-gen-plus/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/users")
	{
		g.GET("", readUsers)         // Get all users
		g.GET("/me", getCurrentUser) // Get current authenticated user
		g.GET("/:id", readUser)      // Get specific user
		g.POST("", createUser)       // Create new user
		g.PATCH("/:id", updateUser)  // Update existing user
		g.DELETE("/:id", deleteUser) // Delete user
	}
}

// readUsers returns all users
func readUsers(c *gin.Context) {
	users, err := core.ReadUsers()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read users")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, users)
}

// readUser returns a specific user by ID
func readUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user ID"})
		return
	}

	user, err := core.ReadUser(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to read user")
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, user)
}

// createUser creates a new user
func createUser(c *gin.Context) {
	var newUser model.User

	if err := c.ShouldBindJSON(&newUser); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind user data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user data"})
		return
	}

	// Validate required fields
	if newUser.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Hash password if provided
	if newUser.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to hash password")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
			return
		}
		newUser.Password = string(hashedPassword)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required for new users"})
		return
	}

	// Get current user from auth
	if !auth.IsLocalAuth() {
		oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
		oauth2Client := c.MustGet("oauth2Client").(auth.Auth)
		currentUser, err := oauth2Client.UserInfo(oauth2Token)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to get current user")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// Use currentUser to avoid compiler error
		log.Debugf("User operation performed by: %s", currentUser.Name)
	}

	createdUser, err := core.CreateUser(&newUser)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create user")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Don't return the password hash
	createdUser.Password = ""

	c.JSON(http.StatusCreated, createdUser)
}

// updateUser updates an existing user
func updateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user ID"})
		return
	}

	var userData model.User
	if err := c.ShouldBindJSON(&userData); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind user data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user data"})
		return
	}

	// Ensure the ID in the URL matches the ID in the body
	userData.Sub = id

	// Get the existing user to check what's changing
	existingUser, err := core.ReadUser(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to read existing user")
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Hash password if provided and different from existing
	if userData.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to hash password")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
			return
		}
		userData.Password = string(hashedPassword)
	} else {
		// If password is not provided, keep the existing one
		// This allows updating other fields without changing password
		userData.Password = existingUser.Password
	}

	// Get current user from auth
	if !auth.IsLocalAuth() {
		oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
		oauth2Client := c.MustGet("oauth2Client").(auth.Auth)
		currentUser, err := oauth2Client.UserInfo(oauth2Token)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to get current user")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// Use currentUser to avoid compiler error
		log.Debugf("User %s is updating user %s", currentUser.Name, id)
	}

	updatedUser, err := core.UpdateUser(id, &userData)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to update user")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Don't return the password hash
	updatedUser.Password = ""

	c.JSON(http.StatusOK, updatedUser)
}

// deleteUser removes a user
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user ID"})
		return
	}

	err := core.DeleteUser(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to delete user")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// getCurrentUser returns the current authenticated user
func getCurrentUser(c *gin.Context) {
	// In local auth mode, we need to get user from session
	if auth.IsLocalAuth() {
		// Get user ID from context
		userID, exists := c.Get("userID")
		if !exists {
			log.Warn("No user ID in context, user may not be authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			log.Error("User ID in context is not a string")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Get user from storage
		user, err := core.ReadUser(userIDStr)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"id":  userIDStr,
			}).Error("Failed to read current user")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Don't return password hash
		user.Password = ""

		// Return user info
		c.JSON(http.StatusOK, user)
		return
	}

	// In OAuth mode, get user info from provider
	oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(auth.Auth)
	userInfo, err := oauth2Client.UserInfo(oauth2Token)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Failed to get user info from OAuth provider")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Try to find user in our database by sub/ID
	user, err := core.ReadUser(userInfo.Sub)
	if err != nil {
		// User might not exist in our database yet
		// For OAuth users, we could create them on the fly
		newUser := &model.User{
			Sub:     userInfo.Sub,
			Name:    userInfo.Name,
			Email:   userInfo.Email,
			IsAdmin: false, // Default to non-admin for new users
		}

		// Create the user if they don't exist
		user, err = core.CreateUser(newUser)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"sub": userInfo.Sub,
			}).Error("Failed to create user from OAuth info")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	// Return user info
	c.JSON(http.StatusOK, user)
}
