package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	provider := flag.String("provider", "google.com", "Auth provider (google.com, facebook.com, apple.com, twitter.com)")
	uid := flag.String("uid", "", "User UID (defaults based on provider)")
	email := flag.String("email", "", "User Email (defaults based on provider)")
	name := flag.String("name", "", "User Name (defaults based on provider)")
	projectID := flag.String("project", "fb-go-commerce-auth", "Firebase Project ID")
	flag.Parse()

	// Defaults based on provider
	if *uid == "" {
		switch *provider {
		case "google.com":
			*uid = "google-uid-123"
		case "facebook.com":
			*uid = "facebook-uid-456"
		case "apple.com":
			*uid = "apple-uid-789"
		case "twitter.com":
			*uid = "twitter-uid-012"
		default:
			*uid = "default-uid"
		}
	}

	if *email == "" {
		switch *provider {
		case "google.com":
			*email = "google-user@example.com"
		case "facebook.com":
			*email = "facebook-user@example.com"
		case "apple.com":
			*email = "apple-user@example.com"
		case "twitter.com":
			*email = "twitter-user@example.com"
		default:
			*email = "user@example.com"
		}
	}

	if *name == "" {
		switch *provider {
		case "google.com":
			*name = "Google User"
		case "facebook.com":
			*name = "Facebook User"
		case "apple.com":
			*name = "Apple User"
		case "twitter.com":
			*name = "Twitter User"
		default:
			*name = "Tester"
		}
	}

	header := map[string]string{
		"alg": "none",
		"typ": "JWT",
	}

	payload := map[string]interface{}{
		"email":          *email,
		"email_verified": true,
		"name":           *name,
		"auth_time":      time.Now().Unix(),
		"user_id":        *uid,
		"firebase": map[string]interface{}{
			"identities": map[string]interface{}{
				"email": []string{*email},
			},
			"sign_in_provider": *provider,
		},
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
		"aud": *projectID,
		"iss": "https://securetoken.google.com/" + *projectID,
		"sub": *uid,
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		log.Fatalf("error marshaling header: %v", err)
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("error marshaling payload: %v", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	fmt.Printf("%s.%s.\n", headerB64, payloadB64)
}
