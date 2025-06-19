package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtSet creates a signed JWT token with any user-defined claim struct and sets it as a cookie.
func JwtSet(w http.ResponseWriter, secure bool, name string, customClaims any, expires time.Time, secret string) error {
	// Marshal user-defined claims to JSON
	claimsBytes, err := json.Marshal(customClaims)
	if err != nil {
		return fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Unmarshal into jwt.MapClaims
	var mapClaims jwt.MapClaims
	if err := json.Unmarshal(claimsBytes, &mapClaims); err != nil {
		return fmt.Errorf("failed to unmarshal to map claims: %w", err)
	}

	// Set exp if not already present
	if _, ok := mapClaims["exp"]; !ok {
		mapClaims["exp"] = expires.Unix()
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return err
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   secure,
		Name:     name,
		Value:    signedToken,
		Expires:  expires,
	})

	return nil
}

// JwtValidate parses and verifies the JWT, and fills the provided struct pointer with claims.
func JwtValidate(r *http.Request, name string, secret string, out any) error {
	cookie, err := r.Cookie(name)
	if err != nil {
		return err
	}

	// Parse token using MapClaims
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Extract and re-marshal MapClaims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		claimsBytes, err := json.Marshal(claims)
		if err != nil {
			return err
		}
		// Fill the output struct
		if err := json.Unmarshal(claimsBytes, out); err != nil {
			return fmt.Errorf("invalid claims structure: %w", err)
		}
		return nil
	}

	return fmt.Errorf("could not extract claims")
}
