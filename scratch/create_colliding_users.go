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

	users := []struct {
		UID   string
		Email string
		Name  string
	}{
		{UID: "john-google", Email: "john@gmail.com", Name: "John Google"},
		{UID: "john-fb", Email: "john@yahoo.com", Name: "John Facebook"},
	}

	for _, user := range users {
		params := (&auth.UserToCreate{}).
			UID(user.UID).
			Email(user.Email).
			DisplayName(user.Name).
			EmailVerified(true)

		u, err := client.CreateUser(context.Background(), params)
		if err != nil {
			log.Printf("error creating user %s: %v", user.UID, err)
		} else {
			log.Printf("Successfully created user: %s (UID: %s)", u.DisplayName, u.UID)
		}
	}
}
