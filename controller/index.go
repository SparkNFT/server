package controller

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	Engine *gin.Engine
)

type ErrorMessage struct {
	Message string `json:"message"`
}

// CORS middleware
func middlewareCors() gin.HandlerFunc {
	cors_config := cors.DefaultConfig()
	cors_config.AllowAllOrigins = true
	cors_config.AllowWildcard = true
	// cors_config.AllowOrigins = []string{CORS_ORIGIN_URL}
	return cors.New(cors_config)
}

// Init initializes controller
func Init() {
	if Engine != nil {
		return
	}

	Engine = gin.Default()
	Engine.Use(middlewareCors())
	Engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	})
	Engine.GET("/api/v1/nft/info", nft_info)
	Engine.GET("/api/v1/nft/list", nft_list)
	Engine.POST("/api/v1/key/claim", claim_key)
}
