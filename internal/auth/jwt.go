package auth

import (
	"context"
	"errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"
	"net/http"
	"strings"
)

// ExtractBearerToken extracts the JWT from the Authorization header
func ExtractBearerToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("missing Authorization header")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid Authorization header format")
	}
	return parts[1], nil
}

// VerifyFirebaseJWT verifies the JWT and returns the Firebase UID
func VerifyFirebaseJWT(r *http.Request) (string, error) {
	token, err := ExtractBearerToken(r)
	if err != nil {
		return "", err
	}
	client := setup.GetAuthClient()
	decoded, err := client.VerifyIDToken(context.Background(), token)
	if err != nil {
		return "", err
	}
	return decoded.UID, nil
}
