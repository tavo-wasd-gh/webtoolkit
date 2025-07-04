package session

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var (
	// Configurable defaults
	MaxSessions = 100_000
	TokenLength = 24
	// Sessions
	sessions   = make(map[string]session)
	sessionsMu sync.RWMutex
)

type session struct {
	csrfTokenHash string
	expires       time.Time
	data          any
}

func StartCleanup(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			cleanupExpiredSessions()
		}
	}()
}

func cleanupExpiredSessions() {
	now := time.Now()

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	for k, s := range sessions {
		if now.After(s.expires) {
			delete(sessions, k)
		}
	}
}

// Creates a new session with the given max age and session data,
// returning a session token and a CSRF token.
//
// The intended usage is to send the session token as an HttpOnly cookie to the client,
// which protects it from JavaScript access. The CSRF token should be sent separately,
// for example embedded in HTML or fetched via an API,
// and included by the client in a header (e.g., "X-CSRF-Token") on subsequent requests
// to validate that the request is legitimate and not forged.
func New(sessionMaxAge time.Duration, sessionData any) (string, string, error) {
    sessionsMu.Lock()
    defer sessionsMu.Unlock()

    if len(sessions) >= MaxSessions {
        return "", "", fmt.Errorf("maximum number of sessions reached")
    }

    st, err := generateToken(TokenLength)
    if err != nil {
        return "", "", fmt.Errorf("error generating session token: %w", err)
    }

    ct, err := generateToken(TokenLength)
    if err != nil {
        return "", "", fmt.Errorf("error generating CSRF token: %w", err)
    }

    hst := hash(st)
    hct := hash(ct)

    sessions[hst] = session{
        csrfTokenHash: hct,
        expires:       time.Now().Add(sessionMaxAge),
        data:          sessionData,
    }

    return st, ct, nil
}

// Validate checks session and CSRF tokens, rotates them if valid, and returns new tokens and session data.
func Validate(st, ct string, dest any) (string, string, error) {
	hst := hash(st)

	sessionsMu.RLock()
	s, ok := sessions[hst]
	sessionsMu.RUnlock()

	if !ok {
		return "", "", fmt.Errorf("invalid session token")
	}

	if time.Now().After(s.expires) {
		sessionsMu.Lock()
		delete(sessions, hst)
		sessionsMu.Unlock()
		return "", "", fmt.Errorf("session expired")
	}

	if subtle.ConstantTimeCompare([]byte(s.csrfTokenHash), []byte(hash(ct))) != 1 {
		return "", "", fmt.Errorf("invalid CSRF token")
	}

	if dest != nil {
		destVal := reflect.ValueOf(dest)
		if destVal.Kind() != reflect.Ptr {
			return "", "", fmt.Errorf("dest must be a pointer")
		}
		dataVal := reflect.ValueOf(s.data)
		if !dataVal.Type().AssignableTo(destVal.Elem().Type()) {
			return "", "", fmt.Errorf("session data type mismatch: expected %T, got %T", destVal.Elem().Interface(), s.data)
		}
		destVal.Elem().Set(dataVal)
	}

	newst, err := generateToken(TokenLength)
	if err != nil {
		return "", "", fmt.Errorf("error generating new session token: %w", err)
	}
	newct, err := generateToken(TokenLength)
	if err != nil {
		return "", "", fmt.Errorf("error generating new CSRF token: %w", err)
	}

	newhst := hash(newst)
	newhct := hash(newct)

	sessionsMu.Lock()
	delete(sessions, hst)
	sessions[newhst] = session{
		csrfTokenHash: newhct,
		expires:       s.expires,
		data:          s.data,
	}
	sessionsMu.Unlock()

	return newst, newct, nil
}

func Delete(st string) error {
    hst := hash(st)

    sessionsMu.Lock()
    defer sessionsMu.Unlock()

    if _, ok := sessions[hst]; !ok {
        return fmt.Errorf("invalid session token")
    }

    delete(sessions, hst)
    return nil
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
