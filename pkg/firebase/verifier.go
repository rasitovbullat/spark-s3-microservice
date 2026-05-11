// Package firebase provides Firebase token verification.
package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// TokenVerifier interface for verifying Firebase ID tokens.
type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, token string) (string, error)
}

// FirebaseVerifier implements TokenVerifier using Firebase Admin SDK.
type FirebaseVerifier struct {
	client *auth.Client
}

// NewFirebaseVerifier creates a new Firebase token verifier.
// credentialsPath is the path to the Firebase service account JSON file.
func NewFirebaseVerifier(credentialsPath string) (*FirebaseVerifier, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase auth client: %w", err)
	}

	return &FirebaseVerifier{client: client}, nil
}

// VerifyIDToken verifies a Firebase ID token and returns the user ID (uid).
func (v *FirebaseVerifier) VerifyIDToken(ctx context.Context, token string) (string, error) {
	decodedToken, err := v.client.VerifyIDToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	return decodedToken.UID, nil
}