package qsite

import "testing"

func TestParsePriorityEncodingOption(t *testing.T) {
	outcomes := []struct {
		input    string
		encoding string
		priority float64
	}{
		{"gzip;q=0.99", "gzip", 0.99},
		{" identity; q=0.39", "identity", 0.39},
		{"deflate ;q=0.42 ", "deflate", 0.42},
		{"deflate;v=9 ", "deflate", 0.0},
	}

	for _, outcome := range outcomes {
		encoding, priority := parsePriorityEncodingOption(outcome.input)

		if encoding != outcome.encoding || priority != outcome.priority {
			t.Fatalf("expected `%s` to return (%s, %f), got (%s, %f)",
				outcome.input,
				outcome.encoding,
				outcome.priority,
				encoding,
				priority,
			)
		}
	}
}

func TestExtractEncodings(t *testing.T) {
	outcomes := []struct {
		input     string
		encodings []string
	}{
		{"identity,gzip,deflate,*", []string{"identity", "gzip", "deflate", "*"}},
		{"gzip, deflate, br, zstd", []string{"gzip", "deflate", "br", "zstd"}},
		{"deflate, gzip;q=1.2", []string{"deflate", "gzip"}},
	}

	for _, outcome := range outcomes {
		encodings := extractEncodings(outcome.input)

		for i, expected := range outcome.encodings {
			if encodings[i] != expected {
				t.Fatalf("expected `%s` to return %+v, got %+v", outcome.input, outcome.encodings, encodings)
			}
		}
	}
}
