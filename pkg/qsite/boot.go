package qsite

import (
	"flag"
)

// BootstrapDefault provides an out-of-the-box entrypoint if you are embedding
// qsite rather than using the published executable. This operation will read
// the CLI flags and then start the server. This is a blocking operation.
func BootstrapDefault() error {
	addr := flag.String("addr", "localhost:8000", "server listen address")
	root := flag.String("root", ".", "content root")
	staticTTL := flag.Int("static-ttl", 600, "static content TTL")
	flag.Parse()

	srv := NewServer(*addr, *root, *staticTTL)
	return srv.Listen()
}
