package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtSet creates a JWT from any claims struct and sets it as a cookie.
func JwtSet[T jwt.Claims](w http.ResponseWriter, secure bool, name string, claims T, expires time.Time, secret string) error {
	// Add expiration to the claims if it's a RegisteredClaims type or contains it
	if rc, ok := any(claims).(*jwt.RegisteredClaims); ok {
		rc.ExpiresAt = jwt.NewNumericDate(expires)
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  expires,
	})

	return nil
}

// JwtValidate reads the cookie, parses the JWT, and populates the provided claims struct.
func JwtValidate[T jwt.Claims](r *http.Request, name string, secret string, claims T) (T, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return claims, err
	}

	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return claims, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return claims, err
	}

	if token.Valid {
		return claims, nil
	}

	return claims, errors.New("invalid token")
}
