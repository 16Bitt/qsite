SERVER_PATH = cmd/qsite/main.go
GORUN = go run $(SERVER_PATH)

.PHONY: example
example:
	$(GORUN) -root ./example/
