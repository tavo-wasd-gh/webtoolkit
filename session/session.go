package session

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	defaultTokenLength    = 24
	defaultSameSitePolicy = http.SameSiteStrictMode
)

var (
	DefaultCookieTTL   = 20 * time.Minute
	DefaultCleanupTime = 5 * time.Minute
	sessions           = make(map[string]session)
	sessionsMu         sync.RWMutex
)

var (
	ErrInvalidSession = errors.New("invalid session token")
	ErrExpiredSession = errors.New("session expired")
	ErrInvalidCSRF    = errors.New("invalid CSRF token")
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

func New(w http.ResponseWriter, secure bool, cookieName string, sessionMaxAge time.Duration, sessionData any) (string, string, error) {
	st, err := generateToken(defaultTokenLength)
	if err != nil {
		return "", "", fmt.Errorf("error generating session token: %w", err)
	}

	ct, err := generateToken(defaultTokenLength)
	if err != nil {
		return "", "", fmt.Errorf("error generating CSRF token: %w", err)
	}

	hst := hash(st)
	hct := hash(ct)

	sessionsMu.Lock()
	sessions[hst] = session{
		csrfTokenHash: hct,
		expires:       time.Now().Add(sessionMaxAge),
		data:          sessionData,
	}
	sessionsMu.Unlock()

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    st,
		MaxAge:   int(DefaultCookieTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: defaultSameSitePolicy,
	}

	http.SetCookie(w, cookie)

	return st, ct, nil
}

func Validate(st, ct string) (string, string, any, error) {
	hst := hash(st)

	sessionsMu.RLock()
	s, ok := sessions[hst]
	sessionsMu.RUnlock()

	if !ok {
		return "", "", nil, ErrInvalidSession
	}

	if time.Now().After(s.expires) {
		sessionsMu.Lock()
		delete(sessions, hst)
		sessionsMu.Unlock()
		return "", "", nil, ErrExpiredSession
	}

	if subtle.ConstantTimeCompare([]byte(s.csrfTokenHash), []byte(hash(ct))) != 1 {
		return "", "", nil, ErrInvalidCSRF
	}

	newst, err := generateToken(defaultTokenLength)
	if err != nil {
		return "", "", nil, fmt.Errorf("error generating new session token: %w", err)
	}
	newct, err := generateToken(defaultTokenLength)
	if err != nil {
		return "", "", nil, fmt.Errorf("error generating new CSRF token: %w", err)
	}

	newhst := hash(newst)
	newhct := hash(newct)

	sessionsMu.Lock()
	delete(sessions, hst)
	sessions[newhst] = session{
		csrfTokenHash: newhct,
		expires:       time.Now().Add(DefaultCookieTTL),
		data:          s.data,
	}
	sessionsMu.Unlock()

	return newst, newct, s.data, nil
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
