package server

import (
	"net/http"
	"wg-gen-plus/auth"
	"wg-gen-plus/core"
	"wg-gen-plus/model"
	"wg-gen-plus/version"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/server")
	{
		g.GET("", readServer)
		g.PATCH("", updateServer)
		g.GET("/config", configServer)
		g.GET("/version", versionStr)
	}
}

func readServer(c *gin.Context) {
	client, err := core.ReadServer()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func updateServer(c *gin.Context) {
	var data model.Server

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	// Support both OAuth2 and local authentication
	if auth.IsLocalAuth() {
		userID, exists := c.Get("userID")
		if !exists {
			log.Error("userID not found in context for local auth")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userIDStr, ok := userID.(string)
		if !ok {
			log.Error("userID in context is not a string")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		data.UpdatedBy = userIDStr
	} else {
		oauth2Token, exists := c.Get("oauth2Token")
		if !exists {
			log.Error("oauth2Token not found in context for OAuth2 auth")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		oauth2Client, exists := c.Get("oauth2Client")
		if !exists {
			log.Error("oauth2Client not found in context for OAuth2 auth")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		user, err := oauth2Client.(auth.Auth).UserInfo(oauth2Token.(*oauth2.Token))
		if err != nil {
			log.WithFields(log.Fields{
				"oauth2Token": oauth2Token,
				"err":         err,
			}).Error("failed to get user with oauth token")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		data.UpdatedBy = user.Name
	}

	server, err := core.UpdateServer(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, server)
}

func configServer(c *gin.Context) {
	configData, err := core.ReadWgConfigFile()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read wg config file")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// return config as txt file
	c.Header("Content-Disposition", "attachment; filename=wg0.conf")
	c.Data(http.StatusOK, "application/config", configData)
}

func versionStr(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version.Version,
	})
}
