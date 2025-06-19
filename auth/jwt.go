package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func JwtSet(w http.ResponseWriter, secure bool, path, name string, customClaims any, expires time.Time, secret string) error {
	claimsBytes, err := json.Marshal(customClaims)
	if err != nil {
		return fmt.Errorf("failed to marshal claims: %w", err)
	}

	var mapClaims jwt.MapClaims
	if err := json.Unmarshal(claimsBytes, &mapClaims); err != nil {
		return fmt.Errorf("failed to unmarshal to map claims: %w", err)
	}

	if _, ok := mapClaims["exp"]; !ok {
		mapClaims["exp"] = expires.Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   secure,
		Name:     name,
		Value:    signedToken,
		Path:     path,
		Expires:  expires,
	})

	return nil
}

func JwtValidate(r *http.Request, path, name, secret string, out any) error {
	cookie, err := r.Cookie(name)
	if err != nil {
		return fmt.Errorf("error validating JWT: %w", err)
	}

	if path != "" && r.URL.Path != path && !pathIsPrefix(path, r.URL.Path) {
		return fmt.Errorf("cookie path mismatch: expected %s, got %s", path, r.URL.Path)
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		claimsBytes, err := json.Marshal(claims)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(claimsBytes, out); err != nil {
			return fmt.Errorf("invalid claims structure: %w", err)
		}
		return nil
	}

	return fmt.Errorf("could not extract claims")
}

func pathIsPrefix(cookiePath, requestPath string) bool {
	return len(requestPath) >= len(cookiePath) && requestPath[:len(cookiePath)] == cookiePath
}
