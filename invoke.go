package main

import (
	"fmt"
	"gsutil/config"
	"gsutil/middleware"
	"gsutil/routes"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type FolderInfo struct {
	FolderPath string `json:"folderPath"`
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
func main() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods:  []string{"GET, POST, PUT, DELETE, OPTIONS"},
		AllowHeaders:  []string{"Origin, X-Requested-With, Content-Type, Accept, Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		AllowOriginFunc: func(origin string) bool {
			allowedOrigins := []string{"http://localhost:3000", os.Getenv("WEB_APP_URL")}
			return contains(allowedOrigins, origin)
		},
	}))
	firebaseAuth := config.SetupFirebase()
	r.Use(func(c *gin.Context) {
		// set firebase auth
		c.Set("firebaseAuth", firebaseAuth)
	})
	r.Use(middleware.AuthMiddleware)
	r.GET("/welcome", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome")
	})
	r.POST("/zip", gin.WrapF(routes.DownloadHandler))
	r.POST("/move", gin.WrapF(routes.MovingObjects))
	r.GET("/list", gin.WrapF(routes.ListDir))
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" //Change this when testing locally
		fmt.Printf("defaulting to port %s", port)
	}
	r.Run(":" + port)
}
