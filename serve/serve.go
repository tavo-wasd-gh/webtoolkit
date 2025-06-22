package serve

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func Compressed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gzw := gzip.NewWriter(w)
		defer gzw.Close()

		gzrw := gzipResponseWriter{Writer: gzw, ResponseWriter: w}
		next.ServeHTTP(gzrw, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
