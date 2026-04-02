package qsite

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

// Stats tracks basic statistics about the content served from the server.
type Stats struct {
	totalHits  *atomic.Uint64
	hitsByPage *sync.Map
}

func NewStats() *Stats {
	return &Stats{
		totalHits:  &atomic.Uint64{},
		hitsByPage: &sync.Map{},
	}
}

// Hit updates the page hit counters.
func (s *Stats) Hit(page string) {
	s.totalHits.Add(1)

	initial := &atomic.Uint64{}
	initial.Store(uint64(1))
	counter, got := s.hitsByPage.LoadOrStore(page, initial)
	if got {
		counter.(*atomic.Uint64).Add(1)
	}
}

func (s *Stats) getHitsByPage() map[string]int {
	result := make(map[string]int)
	s.hitsByPage.Range(func(key, value any) bool {
		result[key.(string)] = int(value.(*atomic.Uint64).Load())
		return true
	})

	return result
}

// ToExposition converts the statistics into the Prometheus exposition format.
func (s *Stats) ToExposition() []byte {
	builder := &bytes.Buffer{}
	hostname, _ := os.Hostname()
	builder.WriteString(`# HELP total_hits Total HTTP pages served
# TYPE total_hits counter
`)
	builder.WriteString(fmt.Sprintf("total_hits{host=\"%s\"} %d\n\n", hostname, s.totalHits.Load()))

	hits := s.getHitsByPage()
	builder.WriteString(`# HELP hits_by_pages HTTP requests per page.
# TYPE hits_by_page counter
`)
	for page, hits := range hits {
		builder.WriteString(fmt.Sprintf("hits_by_page{host=\"%s\",page=\"%s\"} %d\n", hostname, page, hits))
	}

	return builder.Bytes()
}

// Handler returns an HTTP handler that serves the statistics in Prometheus
// exposition format.
func (s *Stats) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Hit("_metrics")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8; version=0.0.4")
		w.Header().Set("Cache-Control", "max-age=0")
		w.Write(s.ToExposition())
	}
}
