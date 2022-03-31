package config

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func SetupFirebase() *storage.Client {
	env, err := LoadConfig("./")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}
	var opt option.ClientOption
	dir, _ := os.Getwd()
	serviceAccountKeyFilePath, err := filepath.Abs(dir + "/" + env.CredentialFile)
	if err != nil {
		panic("Unable to load service account file")
	}
	opt = option.WithCredentialsFile(serviceAccountKeyFilePath)

	clientStorage, err := storage.NewClient(context.Background(), opt)
	if err != nil {
		log.Fatalf("error getting Storage client: %v\n", err)
	}

	return clientStorage
}
