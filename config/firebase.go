package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

func SetupFirebase() *auth.Client {
	var opt option.ClientOption
	serviceAccountKeyFilePath, err := filepath.Abs("../code-red.json")
	fmt.Println("server account:", serviceAccountKeyFilePath)
	if err != nil {
		panic("Unable to load service account file")
	}
	if os.Getenv("ENV") == "production" {
		opt = nil
	} else {
		// THIS SHOULD ONLY BE THERE FOR LOCAL USAGE
		opt = option.WithCredentialsFile(serviceAccountKeyFilePath)
	}
	//Firebase admin SDK initialization
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic("Firebase load error")
	}
	fmt.Println(app)
	//Firebase Auth
	auth, err := app.Auth(context.Background())
	if err != nil {
		panic("Firebase auth load error")
	}

	return auth
}
