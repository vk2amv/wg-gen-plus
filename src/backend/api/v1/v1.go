package apiv1

import (
	"wg-gen-plus/api/v1/auth"
	"wg-gen-plus/api/v1/client"
	"wg-gen-plus/api/v1/server"
	"wg-gen-plus/api/v1/status"
	"wg-gen-plus/api/v1/users"

	"github.com/gin-gonic/gin"
)

// ApplyRoutes apply routes to gin router
func ApplyRoutes(r *gin.RouterGroup, private bool) {
	v1 := r.Group("/v1.0")
	{
		if private {
			client.ApplyRoutes(v1)
			server.ApplyRoutes(v1)
			status.ApplyRoutes(v1)
			users.ApplyRoutes(v1)
		} else {
			auth.ApplyRoutes(v1)
		}
	}
}
