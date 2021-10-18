package config

import (
	"context"
	"fmt"
	"path/filepath"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

func SetupFirebase() *auth.Client {
	serviceAccountKeyFilePath, err := filepath.Abs("../code-red.json")
	fmt.Println("server account:", serviceAccountKeyFilePath)
	if err != nil {
		panic("Unable to load service account file")
	}
	// THIS SHOULD ONLY BE THERE FOR LOCAL USAGE
	//opt := option.WithCredentialsFile(serviceAccountKeyFilePath)
	//Firebase admin SDK initialization
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		panic("Firebase load error")
	}
	fmt.Print("app:", app)
	//Firebase Auth
	auth, err := app.Auth(context.Background())
	if err != nil {
		panic("Firebase load error")
	}
	return auth
}
