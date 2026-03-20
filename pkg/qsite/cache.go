package qsite

import (
	"fmt"
	"net/http"
)

// CacheHandler is a special http.Handler that applies a TTL to served content.
type CacheHandler struct {
	original http.Handler
	cacheTTL int
}

// WrapCache wraps an existing http.Handler in a CacheHandler.
func WrapCache(original http.Handler, ttl int) CacheHandler {
	return CacheHandler{original: original, cacheTTL: ttl}
}

func (ch CacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", ch.cacheTTL))
	ch.original.ServeHTTP(w, r)
}
