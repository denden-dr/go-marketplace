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

	params := (&auth.UserToCreate{}).
		UID("unverified-user").
		Email("unverified@example.com").
		DisplayName("Unverified User").
		EmailVerified(false)

	u, err := client.CreateUser(context.Background(), params)
	if err != nil {
		log.Fatalf("error creating user: %v", err)
	}
	log.Printf("Successfully created user: %s (UID: %s)", u.DisplayName, u.UID)
}
