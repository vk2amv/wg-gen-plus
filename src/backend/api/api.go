package api

import (
	apiv1 "wg-gen-plus/api/v1"

	"github.com/gin-gonic/gin"
)

// ApplyRoutes apply routes to gin engine
func ApplyRoutes(r *gin.Engine, private bool) {
	api := r.Group("/api")
	{
		apiv1.ApplyRoutes(api, private)
	}
}
