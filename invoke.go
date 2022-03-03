package main

import (
	"context"
	"fmt"
	"gsutil/config"
	"gsutil/middleware"
	"gsutil/routes"
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
	appconfig := config.InitEnv()
	_, firebaseStorage := config.SetupFirebase()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"OPTIONS, GET, POST, PUT, DELETE"},
		AllowHeaders:     []string{"Origin, X-Requested-With, Content-Type, Accept, Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			allowedOrigins := []string{"http://localhost:3000", appconfig.WebAppUrl}
			return contains(allowedOrigins, origin)
		},
	}))
	dir, _ := os.Getwd()
	serviceAccountKeyFilePath, err := filepath.Abs(dir + "/" + appconfig.CredentialFile)
	if err != nil {
		panic("Unable to load service account file")
	}
	opt := option.WithCredentialsFile(serviceAccountKeyFilePath)

	client, err := storage.NewClient(context.Background(), opt)
	if err != nil {
		fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	routes.InitStorageClient(appconfig.StorageBucket, firebaseStorage)
	r.Use(func(c *gin.Context) {
		c.Set("firebaseStorage", firebaseStorage)
	})
	r.Use(middleware.AuthMiddleware)
	r.POST("/zip", gin.WrapF(routes.DownloadHandler))
	r.POST("/move", routes.MovingFiles)
	r.GET("/list/:path", routes.ListObjectsInFolder)
	r.GET("/list", routes.ListObjects)
	r.GET("/list/file/:path", routes.GetObject)
	r.POST("/delete-files", routes.DeleteObjects)
	r.POST("/upload", routes.UploadFiles)
	r.POST("/upload/avatar", routes.UploadAvatar)
	r.POST("/upload/cert", routes.UploadCertBadge)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" //Change this when testing locally
		fmt.Printf("defaulting to port %s", port)
	}
	r.Run(":" + port)
}
