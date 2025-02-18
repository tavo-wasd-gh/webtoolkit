package cors

import (
	"net/http"
	"strings"
)

func valid(given string, allowed string) bool {
	if given == "" || allowed == "" {
		return false
	}

	for _, header := range strings.Split(allowed, ",") {
		if strings.TrimSpace(header) == given {
			return true
		}
	}

	return false
}

func setHeaders(w http.ResponseWriter, origins string, methods string, headers string, credentials bool) {
	w.Header().Set("Access-Control-Allow-Origin", origins)
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", headers)

	if credentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else {
		w.Header().Set("Access-Control-Allow-Credentials", "false")
	}
}

func Handler(w http.ResponseWriter, r *http.Request, origins string, methods string, headers string, credentials bool) bool {
	if r.Method == http.MethodOptions {
		setHeaders(w, origins, methods, headers, credentials)

		if valid(r.Header.Get("Origin"), origins) || origins == "*" {
			w.WriteHeader(http.StatusOK)
			return false
		}

		w.WriteHeader(http.StatusForbidden)
		return false
	}

	if !valid(r.Method, methods) {
		setHeaders(w, origins, methods, headers, credentials)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}

	if !valid(r.Header.Get("Origin"), origins) && origins != "*" {
		setHeaders(w, origins, methods, headers, credentials)
		w.WriteHeader(http.StatusForbidden)
		return false
	}

	setHeaders(w, origins, methods, headers, credentials)
	return true
}
