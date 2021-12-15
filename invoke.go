package main

import (
	"context"
	"fmt"
	"gsutil/config"
	"gsutil/middleware"
	"gsutil/routes"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
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
	env, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}
	dir, _ := os.Getwd()
	serviceAccountKeyFilePath, err := filepath.Abs(dir + "/" + env.CredentialFile)
	if err != nil {
		panic("Unable to load service account file")
	}
	opt := option.WithCredentialsFile(serviceAccountKeyFilePath)

	client, err := storage.NewClient(context.Background(), opt)
	if err != nil {
		fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()
	routes.Init(env.StorageBucket, client)
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"OPTIONS, GET, POST, PUT, DELETE"},
		AllowHeaders:     []string{"Origin, X-Requested-With, Content-Type, Accept, Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		AllowAllOrigins:  true,
		/* AllowOriginFunc: func(origin string) bool {
			allowedOrigins := []string{"http://localhost:3000", env.WebAppUrl}
			return contains(allowedOrigins, origin)
		}, */
	}))
	firebaseAuth, firebaseStorage := config.SetupFirebase()
	r.Use(func(c *gin.Context) {
		// set firebase auth
		c.Set("firebaseAuth", firebaseAuth)
		c.Set("firebaseStorage", firebaseStorage)
	})
	r.Use(middleware.AuthMiddleware)
	//r.Use(middleware.GinBodyWriter)
	r.POST("/zip", gin.WrapF(routes.DownloadHandler))
	r.POST("/move", gin.WrapF(routes.MovingObjects))
	r.GET("/list/:path", routes.ListObjects)
	/* r.POST("/delete-files", gin.WrapF(routes.DeleteObjects)) */
	r.POST("/delete-files", routes.DeleteObjects)
	r.POST("/upload", routes.UploadFiles)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" //Change this when testing locally
		fmt.Printf("defaulting to port %s", port)
	}
	r.Run(":" + port)
}
