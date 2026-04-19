//go:build ignore
// +build ignore

package main

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type TestUser struct {
	UID         string
	Email       string
	DisplayName string
}

func main() {
	os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099")

	config := &firebase.Config{ProjectID: "fb-go-commerce-auth"}
	app, err := firebase.NewApp(context.Background(), config, option.WithoutAuthentication())
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	users := []TestUser{
		{UID: "google-uid-123", Email: "google-user@example.com", DisplayName: "Google User"},
		{UID: "facebook-uid-456", Email: "facebook-user@example.com", DisplayName: "Facebook User"},
		{UID: "apple-uid-789", Email: "apple-user@example.com", DisplayName: "Apple User"},
		{UID: "twitter-uid-012", Email: "twitter-user@example.com", DisplayName: "Twitter User"},
	}

	for _, user := range users {
		params := (&auth.UserToCreate{}).
			UID(user.UID).
			Email(user.Email).
			DisplayName(user.DisplayName).
			EmailVerified(true)

		u, err := client.CreateUser(context.Background(), params)
		if err != nil {
			log.Printf("User %s (%s) already exists or failed: %v", user.DisplayName, user.UID, err)
		} else {
			log.Printf("Successfully created user: %s (UID: %s)", u.DisplayName, u.UID)
		}
	}
}
