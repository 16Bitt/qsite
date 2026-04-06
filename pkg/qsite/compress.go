package qsite

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// CompressionHandler is an http.Handler that can be used in cases where
// wrapping individual http.HandlerFunc functions is not possible.
type CompressionHandler struct {
	wrapped http.HandlerFunc
}

// WrapCompression returns a new http.Handler that will apply optional gzip
// compression to responses.
func WrapCompression(handler http.Handler) http.Handler {
	return CompressionHandler{wrapped: MaybeCompress(handler.ServeHTTP)}
}

func (ch CompressionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch.wrapped(w, r)
}

type compressedResponseWriter struct {
	original   http.ResponseWriter
	compressed io.Writer
}

func (c *compressedResponseWriter) Header() http.Header {
	return c.original.Header()
}

func (c *compressedResponseWriter) Write(data []byte) (int, error) {
	return c.compressed.Write(data)
}

func (c *compressedResponseWriter) WriteHeader(status int) {
	c.Header().Del("Content-Length")
	c.original.WriteHeader(status)
}

func MaybeCompress(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !shouldGzip(r) {
			fn(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		defer gw.Close()

		wrapped := &compressedResponseWriter{original: w, compressed: gw}
		fn(wrapped, r)
	}
}

func shouldGzip(r *http.Request) bool {
	encodings := extractEncodings(r.Header.Get("Accept-Encoding"))
	for _, encoding := range encodings {
		if encoding == "gzip" {
			return true
		}
	}

	return false
}

// Parses an Accept-Encoding header string and returns the encodings that are
// not explicitly disabled.
//
// See https://httpwg.org/specs/rfc9110.html#field.accept-encoding
// TODO: respect priority ordering.
func extractEncodings(acceptEncoding string) []string {
	result := make([]string, 0, 1)

	for option := range strings.SplitSeq(acceptEncoding, ",") {
		option = strings.TrimSpace(option)

		if strings.ContainsRune(option, ';') {
			option, priority := parsePriorityEncodingOption(option)
			if priority > 0.0 {
				result = append(result, option)
			}
		} else {
			result = append(result, option)
		}
	}

	return result
}

// parseEncodingOption parses a Content-Encoding option in the form "foo;q=0.1"
// and returns encoding name and the priority. If parsing fails, a best effort
// encoding will be returned, with the priority set to 0.0 (meaning: don't use
// this encoding).
//
// See https://httpwg.org/specs/rfc9110.html#field.accept-encoding
func parsePriorityEncodingOption(option string) (string, float64) {
	allParts := strings.SplitN(option, ";", 2)
	if len(allParts) != 2 {
		return option, 0.0
	}
	encoding := strings.TrimSpace(allParts[0])

	prioParts := strings.SplitN(allParts[1], "=", 2)
	if len(prioParts) != 2 {
		return encoding, 0.0
	}
	key := strings.TrimSpace(prioParts[0])
	if key != "q" {
		return encoding, 0.0
	}
	priority, err := strconv.ParseFloat(strings.TrimSpace(prioParts[1]), 64)
	if err != nil {
		return encoding, 0.0
	}

	return encoding, priority
}
