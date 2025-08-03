package client

import (
	"net/http"

	"wg-gen-plus/auth"
	"wg-gen-plus/core"
	"wg-gen-plus/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/client")
	{
		g.POST("", createClient)
		g.GET("/:id", readClient)
		g.PATCH("/:id", updateClient)
		g.DELETE("/:id", deleteClient)
		g.GET("", readClients)
		g.GET("/:id/config", configClient)
		g.GET("/:id/email", emailClient)
	}
}

func createClient(c *gin.Context) {
	var data model.Client

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	// Get creator info based on authentication type
	var createdBy string

	if auth.IsLocalAuth() {
		// Local auth: Get user from context
		userID, exists := c.Get("userID")
		if !exists {
			log.Error("User ID not found in context")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			log.Error("User ID is not a string")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Get the user from database
		user, err := core.ReadUser(userIDStr)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"id":  userIDStr,
			}).Error("failed to get user")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		createdBy = user.Name
	} else {
		// OAuth: Get user from token
		oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
		oauth2Client := c.MustGet("oauth2Client").(auth.Auth)
		user, err := oauth2Client.UserInfo(oauth2Token)
		if err != nil {
			log.WithFields(log.Fields{
				"oauth2Token": oauth2Token,
				"err":         err,
			}).Error("failed to get user with oauth token")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		createdBy = user.Name
	}

	data.CreatedBy = createdBy

	client, err := core.CreateClient(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func readClient(c *gin.Context) {
	id := c.Param("id")

	client, err := core.ReadClient(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func updateClient(c *gin.Context) {
	var data model.Client
	id := c.Param("id")

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	// Get updater info based on authentication type
	var updatedBy string

	if auth.IsLocalAuth() {
		// Local auth: Get user from context
		userID, exists := c.Get("userID")
		if !exists {
			log.Error("User ID not found in context")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			log.Error("User ID is not a string")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Get the user from database
		user, err := core.ReadUser(userIDStr)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"id":  userIDStr,
			}).Error("failed to get user")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		updatedBy = user.Name
		log.Debugf("User %s is updating client %s", updatedBy, id)
	} else {
		// OAuth: Get user from token
		oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
		oauth2Client := c.MustGet("oauth2Client").(auth.Auth)
		user, err := oauth2Client.UserInfo(oauth2Token)
		if err != nil {
			log.WithFields(log.Fields{
				"oauth2Token": oauth2Token,
				"err":         err,
			}).Error("failed to get user with oauth token")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		updatedBy = user.Name
	}

	data.UpdatedBy = updatedBy

	client, err := core.UpdateClient(id, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func deleteClient(c *gin.Context) {
	id := c.Param("id")

	err := core.DeleteClient(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func readClients(c *gin.Context) {
	clients, err := core.ReadClients()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, clients)
}

func configClient(c *gin.Context) {
	configData, err := core.ReadClientConfig(c.Param("id"))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client config")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	formatQr := c.DefaultQuery("qrcode", "false")
	if formatQr == "false" {
		// return config as txt file
		c.Header("Content-Disposition", "attachment; filename=wg0.conf")
		c.Data(http.StatusOK, "application/config", configData)
		return
	}
	// return config as png qrcode
	png, err := qrcode.Encode(string(configData), qrcode.Medium, 250)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create qrcode")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "image/png", png)
}

func emailClient(c *gin.Context) {
	id := c.Param("id")

	err := core.EmailClient(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to send email to client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
